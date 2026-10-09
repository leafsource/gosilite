// STable 通用查询表格类型定义
import type { TableProps, TreeOptionData } from 'tdesign-vue-next';

export type STableColumn = NonNullable<TableProps['columns']>[number];

/** 树形表格配置，结构对齐 TdEnhancedTableProps['tree'] */
export interface STableTreeConfig {
  childrenKey?: string;
  treeNodeColumnIndex?: number;
  defaultExpandAll?: boolean;
}

export interface STableSearchField {
  /** 字段名，对应 searchParams 的 key */
  name: string;
  label?: string;
  type?: 'input' | 'select' | 'date' | 'daterange' | 'custom' | 'tree-select';
  /** placeholder；daterange 时可传 [起, 止] */
  placeholder?: string | string[];
  options?: Array<{ label: string; value: string | number }>;
  /** type=tree-select 时的树数据 */
  treeData?: TreeOptionData[];
}

export interface STableProps {
  /** 行主键 */
  rowKey?: string;
  /** 列配置 */
  columns: STableColumn[];
  /** 数据 */
  data: Array<Record<string, unknown>>;
  /** 总条数 */
  total?: number;
  /** 页码 */
  page?: number;
  /** 每页条数 */
  pageSize?: number;
  loading?: boolean;
  /** 快捷关键字占位 */
  keywordPlaceholder?: string;
  /** 是否显示搜索区（默认 true；纯展示表可关） */
  showSearch?: boolean;
  /** 结构化搜索项（渲染在关键字旁） */
  searchFields?: STableSearchField[];
  /** 显示默认列（列 key 集合）；不传则全显示 */
  defaultVisibleKeys?: string[];
  /** 是否显示选择框（默认显示，配合 #batch 插槽做批量操作） */
  selectable?: boolean;
  /** 树形表格配置；传入时内层表格切换为 EnhancedTable */
  tree?: STableTreeConfig;
  /** 是否显示分页条（全量数据页可关），默认 true */
  paginated?: boolean;
  /** 搜索项初始值 / 重置后的恢复值（如默认时间范围） */
  defaultParams?: Record<string, unknown>;
}

export interface STableEmits {
  (e: 'search', params: Record<string, unknown>): void;
  (e: 'reset'): void;
  (e: 'refresh'): void;
  (e: 'page-change', page: number, pageSize: number): void;
  (e: 'selection-change', selected: Array<Record<string, unknown>>): void;
  (e: 'update:page', page: number): void;
  (e: 'update:pageSize', pageSize: number): void;
}
