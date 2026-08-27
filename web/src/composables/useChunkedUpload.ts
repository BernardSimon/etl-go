import { ref } from 'vue';
import {
    CHUNK_SIZE,
    sha256Hex,
    initUploadSession,
    uploadChunk,
    completeUploadSession,
    cancelUploadSession,
    getUploadStatus,
} from '../api/file';
import type { FileInfo } from '../types';

export type ChunkedUploadState =
    | 'idle'
    | 'hashing'
    | 'initialising'
    | 'resuming'
    | 'uploading'
    | 'completing'
    | 'done'
    | 'paused'
    | 'cancelled'
    | 'error';

const MAX_RETRIES = 3;
const RETRY_DELAYS_MS = [1000, 2000, 4000];

function sleep(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
}

/**
 * Composable that handles large-file chunked uploads with pause/resume/cancel
 * and automatic retry on network errors.
 *
 * Usage:
 *   const upload = useChunkedUpload();
 *   const fileInfo = await upload.start(file, percent => { ... });
 */
export function useChunkedUpload() {
    const state = ref<ChunkedUploadState>('idle');
    const progress = ref(0);
    const currentChunk = ref(0);
    const totalChunks = ref(0);
    const sessionId = ref<string | null>(null);
    const errorMsg = ref<string | null>(null);

    let abortController: AbortController | null = null;
    let isPaused = false;
    let isCancelled = false;
    let pauseResolve: (() => void) | null = null;
    let pausePromise: Promise<void> | null = null;

    /** Pause the in-progress upload after the current chunk finishes. */
    const pause = () => {
        if (state.value !== 'uploading') return;
        isPaused = true;
        state.value = 'paused';
        pausePromise = new Promise(resolve => {
            pauseResolve = resolve;
        });
    };

    /** Resume a paused upload. */
    const resume = () => {
        if (state.value !== 'paused') return;
        isPaused = false;
        state.value = 'uploading';
        pauseResolve?.();
        pauseResolve = null;
        pausePromise = null;
    };

    /** Cancel the upload and clean up the server-side session. */
    const cancel = async () => {
        isCancelled = true;
        abortController?.abort();
        // Unblock any paused await
        if (pauseResolve) {
            pauseResolve();
            pauseResolve = null;
            pausePromise = null;
        }
        isPaused = false;
        state.value = 'cancelled';
        if (sessionId.value) {
            try { await cancelUploadSession(sessionId.value); } catch { /* ignore */ }
        }
    };

    /**
     * Start (or resume) a chunked upload.
     *
     * @param file            The File object to upload.
     * @param onProgress      Optional callback receiving overall percent (0-100).
     * @param resumeSessionId Existing session ID to resume instead of starting fresh.
     */
    const start = async (
        file: File,
        onProgress?: (percent: number) => void,
        resumeSessionId?: string,
    ): Promise<FileInfo> => {
        errorMsg.value = null;
        abortController = new AbortController();
        isPaused = false;
        isCancelled = false;

        const total = Math.ceil(file.size / CHUNK_SIZE);
        totalChunks.value = total;
        currentChunk.value = 0;
        progress.value = 0;

        const receivedSet = new Set<number>();

        // ── Resume path ────────────────────────────────────────────────────
        if (resumeSessionId) {
            state.value = 'resuming';
            sessionId.value = resumeSessionId;
            try {
                const res = await getUploadStatus(resumeSessionId);
                if (res.code === 0) {
                    const s = res.data;
                    if (s.status === 'done') {
                        state.value = 'done';
                        progress.value = 100;
                        // Return a minimal stub — caller should refresh its file list
                        return { id: s.file_id } as FileInfo;
                    }
                    s.received_map.forEach(idx => receivedSet.add(idx));
                    currentChunk.value = receivedSet.size;
                    progress.value = Math.round((receivedSet.size / total) * 100);
                    onProgress?.(progress.value);
                }
            } catch {
                // Status fetch failed — upload all chunks from scratch
            }
        } else {
            // ── Fresh upload: hash whole file then init session ────────────
            state.value = 'hashing';
            let fileHash = '';
            try {
                fileHash = await sha256Hex(file);
            } catch {
                // Hash computation failed — proceed without whole-file verification
            }

            state.value = 'initialising';
            const initRes = await initUploadSession({
                filename: file.name,
                total_size: file.size,
                chunk_size: CHUNK_SIZE,
                total_chunks: total,
                expected_hash: fileHash,
            });
            if (initRes.code !== 0) {
                state.value = 'error';
                errorMsg.value = initRes.message || 'Failed to initialise upload';
                throw new Error(errorMsg.value);
            }
            sessionId.value = initRes.data.session_id;
            totalChunks.value = initRes.data.total_chunks;
        }

        // ── Upload loop ────────────────────────────────────────────────────
        state.value = 'uploading';

        for (let i = 0; i < total; i++) {
            // Skip already-received chunks (resume)
            if (receivedSet.has(i)) continue;

            // Wait while paused
            if (isPaused && pausePromise) await pausePromise;

            // Bail if cancelled (isCancelled avoids TS control-flow narrowing on state.value)
            if (isCancelled) throw new Error('Upload cancelled');

            const start = i * CHUNK_SIZE;
            const blob = file.slice(start, start + CHUNK_SIZE);

            let chunkHash = '';
            try { chunkHash = await sha256Hex(blob); } catch { /* skip */ }

            let uploaded = false;
            for (let attempt = 0; attempt < MAX_RETRIES; attempt++) {
                try {
                    await uploadChunk(sessionId.value!, i, blob, chunkHash, {
                        signal: abortController.signal,
                        onUploadProgress: (event) => {
                            if (!event.total) return;
                            const chunkFrac = event.loaded / event.total;
                            const overall = Math.min(99, Math.round(((i + chunkFrac) / total) * 100));
                            progress.value = overall;
                            onProgress?.(overall);
                        },
                    });
                    uploaded = true;
                    break;
                } catch (err: any) {
                    if (err?.name === 'CanceledError' || err?.code === 'ERR_CANCELED') throw err;
                    if (attempt < MAX_RETRIES - 1) await sleep(RETRY_DELAYS_MS[attempt]);
                }
            }

            if (!uploaded) {
                state.value = 'error';
                errorMsg.value = `Failed to upload chunk ${i} after ${MAX_RETRIES} attempts`;
                throw new Error(errorMsg.value);
            }

            receivedSet.add(i);
            currentChunk.value = receivedSet.size;
        }

        // ── Trigger assembly ───────────────────────────────────────────────
        state.value = 'completing';
        const completeRes = await completeUploadSession(sessionId.value!);
        if (completeRes.code !== 0) {
            state.value = 'error';
            errorMsg.value = completeRes.message || 'Failed to assemble file';
            throw new Error(errorMsg.value);
        }

        state.value = 'done';
        progress.value = 100;
        onProgress?.(100);
        return completeRes.data;
    };

    return {
        state,
        progress,
        currentChunk,
        totalChunks,
        sessionId,
        errorMsg,
        start,
        pause,
        resume,
        cancel,
    };
}
