import { computed, onBeforeUnmount, onMounted, ref } from "vue";

export const MOBILE_BREAKPOINT = 768;
export const DESKTOP_BREAKPOINT = 1200;

export const useResponsive = () => {
  const width = ref(typeof window === "undefined" ? DESKTOP_BREAKPOINT : window.innerWidth);

  const handleResize = () => {
    width.value = window.innerWidth;
  };

  onMounted(() => {
    handleResize();
    window.addEventListener("resize", handleResize);
  });

  onBeforeUnmount(() => {
    window.removeEventListener("resize", handleResize);
  });

  const isMobile = computed(() => width.value < MOBILE_BREAKPOINT);
  const isTablet = computed(() => width.value >= MOBILE_BREAKPOINT && width.value < DESKTOP_BREAKPOINT);
  const isDesktop = computed(() => width.value >= DESKTOP_BREAKPOINT);

  return {
    width,
    isMobile,
    isTablet,
    isDesktop,
  };
};
