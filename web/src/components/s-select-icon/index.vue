<template>
  <t-popup
    v-model:visible="visible"
    trigger="click"
    placement="bottom-left"
    :disabled="disabled"
    :overlay-inner-style="{ padding: 0 }"
  >
    <t-input
      :model-value="modelValue"
      :placeholder="placeholder"
      :disabled="disabled"
      readonly
      class="s-select-icon__trigger"
    >
      <template #prefix-icon>
        <t-icon v-if="modelValue && KNOWN_ICON_SET.has(modelValue)" :name="modelValue" class="s-select-icon__preview" />
      </template>
      <template #suffix>
        <span v-if="clearable && modelValue && !disabled" class="s-select-icon__clear" @click.stop="onClear">
          <t-icon name="close-circle-filled" />
        </span>
        <t-icon name="chevron-down" class="s-select-icon__arrow" />
      </template>
    </t-input>

    <template #content>
      <div v-if="visible" class="s-select-icon">
        <t-input
          v-model="query"
          autofocus
          clearable
          placeholder="搜索图标名称，如 user"
          class="s-select-icon__search"
        />
        <div class="s-select-icon__body">
          <div class="s-select-icon__groups">
            <button
              v-for="g in groups"
              :key="g.label"
              type="button"
              tabindex="-1"
              class="s-select-icon__group"
              :class="{ 'is-active': !query && g.label === currentGroup }"
              @click="onSelectGroup(g.label)"
            >
              <span class="s-select-icon__group-label">{{ g.label }}</span>
              <span class="s-select-icon__group-count">{{ g.icons.length }}</span>
            </button>
          </div>
          <div class="s-select-icon__main">
            <div class="s-select-icon__meta">
              <template v-if="query"> 搜索结果 · {{ filtered.length }} 个 </template>
              <template v-else-if="shownList.length < filtered.length">
                已显示 {{ shownList.length }} / {{ filtered.length }}，滚动加载更多
              </template>
              <template v-else>{{ currentGroup }} · 共 {{ filtered.length }} 个</template>
            </div>
            <div v-if="filtered.length" class="s-select-icon__grid" @scroll="onGridScroll">
              <button
                v-for="stem in shownList"
                :key="stem"
                type="button"
                tabindex="-1"
                class="s-select-icon__cell"
                :class="{ 'is-active': stem === modelValue }"
                :title="stem"
                @click="onSelect(stem)"
              >
                <t-icon :name="stem" />
                <span class="s-select-icon__cell-name">{{ stem }}</span>
              </button>
            </div>
            <div v-else class="s-select-icon__empty">无匹配图标，换个关键词试试</div>
          </div>
        </div>
      </div>
    </template>
  </t-popup>
</template>
<script setup lang="ts">
import { manifest } from 'tdesign-icons-vue-next';
import { computed, ref, watch } from 'vue';

import type { IconGroup } from './iconGroups';
import { ICON_GROUPS } from './iconGroups';

defineOptions({ name: 'SSelectIcon' });

withDefaults(
  defineProps<{
    /** 图标名（tdesign 图标 stem，如 user、folder） */
    modelValue?: string;
    placeholder?: string;
    clearable?: boolean;
    disabled?: boolean;
  }>(),
  { modelValue: '', placeholder: '选择图标', clearable: true, disabled: false },
);

const emit = defineEmits<{
  (e: 'update:model-value', v: string): void;
  (e: 'change', v: string): void;
}>();

const ALL_ICONS: string[] = manifest.map((item) => item.stem);
const KNOWN_ICON_SET = new Set(ALL_ICONS);
const PAGE_SIZE = 240;
const GROUP_ALL = '全部';
const GROUP_OTHER = '其他';

/** 分组列表：全部 + 官方分类（npm 包里的新图标运行时归入「其他」） */
function buildGroups(): IconGroup[] {
  const covered = new Set(ICON_GROUPS.flatMap((g) => g.icons));
  const others = ALL_ICONS.filter((stem) => !covered.has(stem));
  const list: IconGroup[] = [{ label: GROUP_ALL, icons: ALL_ICONS }];
  let hasOther = false;
  for (const g of ICON_GROUPS) {
    if (g.label === GROUP_OTHER) {
      list.push({ label: g.label, icons: [...g.icons, ...others] });
      hasOther = true;
    } else {
      list.push(g);
    }
  }
  if (!hasOther && others.length) list.push({ label: GROUP_OTHER, icons: others });
  return list;
}

const groups = buildGroups();

const visible = ref(false);
const query = ref('');
const currentGroup = ref(GROUP_ALL);
const visibleCount = ref(PAGE_SIZE);

const filtered = computed(() => {
  const q = query.value.trim().toLowerCase();
  if (q) return ALL_ICONS.filter((stem) => stem.includes(q));
  return groups.find((g) => g.label === currentGroup.value)?.icons ?? ALL_ICONS;
});
const shownList = computed(() => filtered.value.slice(0, visibleCount.value));

// 每次打开重置搜索/分组/分页；搜索或切组时重置分页
watch(visible, (v) => {
  if (v) {
    query.value = '';
    currentGroup.value = GROUP_ALL;
    visibleCount.value = PAGE_SIZE;
  }
});
watch([query, currentGroup], () => {
  visibleCount.value = PAGE_SIZE;
});

function onSelectGroup(label: string) {
  currentGroup.value = label;
  query.value = '';
}

function onGridScroll(e: Event) {
  const el = e.currentTarget as HTMLElement;
  if (visibleCount.value < filtered.value.length && el.scrollHeight - el.scrollTop - el.clientHeight < 160) {
    visibleCount.value += PAGE_SIZE;
  }
}

function onSelect(stem: string) {
  emit('update:model-value', stem);
  emit('change', stem);
  visible.value = false;
}

function onClear() {
  emit('update:model-value', '');
  emit('change', '');
}
</script>
<style lang="less" scoped>
.s-select-icon__trigger {
  cursor: pointer;

  :deep(.t-input__inner) {
    cursor: pointer;
  }
}

.s-select-icon__preview {
  color: var(--td-brand-color);
}

.s-select-icon__clear {
  display: inline-flex;
  align-items: center;
  color: var(--td-text-color-placeholder);
  cursor: pointer;

  &:hover {
    color: var(--td-text-color-secondary);
  }
}

.s-select-icon__arrow {
  color: var(--td-text-color-placeholder);
}

.s-select-icon {
  display: flex;
  flex-direction: column;
  width: 480px;
  padding: 12px;
}

.s-select-icon__body {
  display: flex;
  gap: 8px;
  align-items: flex-start;
  margin-top: 8px;
}

.s-select-icon__groups {
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
  gap: 2px;
  width: 104px;
  max-height: 328px;
  padding-right: 4px;
  overflow-y: auto;
  border-right: 1px solid var(--td-border-level-1-color);
}

.s-select-icon__group {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 30px;
  padding: 0 8px;
  color: var(--td-text-color-secondary);
  font-family: inherit;
  font-size: var(--td-font-size-body-small);
  background: transparent;
  border: none;
  border-radius: var(--td-radius-small);
  cursor: pointer;

  &:hover {
    background: var(--td-bg-color-container-hover);
    color: var(--td-text-color-primary);
  }

  &.is-active {
    color: var(--td-brand-color);
    font-weight: 500;
    background: var(--td-brand-color-light);
  }
}

.s-select-icon__group-count {
  color: var(--td-text-color-placeholder);
  font-size: 11px;
  font-variant-numeric: tabular-nums;
}

.s-select-icon__group.is-active .s-select-icon__group-count {
  color: var(--td-brand-color);
}

.s-select-icon__main {
  flex: 1;
  min-width: 0;
}

.s-select-icon__meta {
  color: var(--td-text-color-placeholder);
  font-size: var(--td-font-size-body-small);
  line-height: 20px;
}

.s-select-icon__grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(64px, 1fr));
  gap: 4px;
  max-height: 300px;
  margin-top: 8px;
  overflow-y: auto;
}

.s-select-icon__cell {
  display: flex;
  flex-direction: column;
  gap: 4px;
  align-items: center;
  justify-content: center;
  height: 68px;
  padding: 4px;
  color: var(--td-text-color-primary);
  font-family: inherit;
  background: transparent;
  border: 1px solid transparent;
  border-radius: var(--td-radius-small);
  cursor: pointer;

  &:hover {
    background: var(--td-bg-color-container-hover);
    border-color: var(--td-border-level-2-color);
  }

  &.is-active {
    color: var(--td-brand-color);
    background: var(--td-brand-color-light);
    border-color: var(--td-brand-color);
  }

  .t-icon {
    font-size: 20px;
  }
}

.s-select-icon__cell-name {
  width: 100%;
  overflow: hidden;
  color: var(--td-text-color-secondary);
  font-size: 10px;
  line-height: 14px;
  text-align: center;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.s-select-icon__cell.is-active .s-select-icon__cell-name {
  color: var(--td-brand-color);
}

.s-select-icon__empty {
  padding: 24px 0;
  color: var(--td-text-color-placeholder);
  font-size: var(--td-font-size-body-small);
  text-align: center;
}
</style>
