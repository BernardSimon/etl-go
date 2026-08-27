import { computed } from "vue";
import { useRoute } from "vue-router";
import { useI18n } from "vue-i18n";

export const usePageMeta = () => {
  const route = useRoute();
  const { t } = useI18n();

  const title = computed(() => {
    const value = route.meta.title as string | undefined;
    return value ? t(value) : "";
  });

  const description = computed(() => {
    const value = route.meta.description as string | undefined;
    return value ? t(value) : "";
  });

  return {
    title,
    description,
  };
};
