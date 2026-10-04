<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import {
  MessagePlugin,
  type TableProps,
} from 'tdesign-vue-next'
import {
  AddIcon,
  DeleteIcon,
  EditIcon,
  RefreshIcon,
} from 'tdesign-icons-vue-next'
import {
  createTask,
  deleteTask,
  listTasks,
  updateTask,
  ApiError,
} from './api/tasks'
import type { Task, TaskStatus } from './types/task'

const STATUS_OPTIONS = [
  { label: '全部', value: '' },
  { label: '待办', value: 'todo' },
  { label: '进行中', value: 'doing' },
  { label: '完成', value: 'done' },
] as const

const STATUS_TAG: Record<TaskStatus, { theme: string; label: string }> = {
  todo: { theme: 'default', label: '待办' },
  doing: { theme: 'warning', label: '进行中' },
  done: { theme: 'success', label: '完成' },
}

const tasks = ref<Task[]>([])
const loading = ref(false)
const filterStatus = ref('')
const createVisible = ref(false)
const editVisible = ref(false)
const saving = ref(false)

const createForm = reactive({
  title: '',
  description: '',
  status: 'todo' as TaskStatus,
})

const editForm = reactive({
  id: '',
  title: '',
  description: '',
  status: 'todo' as TaskStatus,
})

const filteredCount = computed(() => tasks.value.length)

const columns: TableProps['columns'] = [
  {
    colKey: 'title',
    title: '标题',
    width: 180,
  },
  {
    colKey: 'description',
    title: '描述',
    ellipsis: true,
  },
  {
    colKey: 'status',
    title: '状态',
    width: 110,
  },
  {
    colKey: 'updated_at',
    title: '更新时间',
    width: 180,
  },
  {
    colKey: 'ops',
    title: '操作',
    width: 150,
    fixed: 'right',
  },
]

function formatTime(iso: string): string {
  if (!iso) return ''
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso
  return d.toLocaleString('zh-CN', { hour12: false })
}

async function reload() {
  loading.value = true
  try {
    tasks.value = await listTasks(filterStatus.value || undefined)
  } catch (e) {
    MessagePlugin.error(e instanceof ApiError ? e.message : '加载失败')
  } finally {
    loading.value = false
  }
}

function openCreate() {
  createForm.title = ''
  createForm.description = ''
  createForm.status = 'todo'
  createVisible.value = true
}

function openEdit(task: Task) {
  editForm.id = task.id
  editForm.title = task.title
  editForm.description = task.description
  editForm.status = task.status
  editVisible.value = true
}

async function submitCreate() {
  if (!createForm.title.trim()) {
    MessagePlugin.warning('请填写标题')
    return
  }
  saving.value = true
  try {
    await createTask({
      title: createForm.title.trim(),
      description: createForm.description.trim(),
      status: createForm.status,
    })
    createVisible.value = false
    MessagePlugin.success('已创建')
    await reload()
  } catch (e) {
    MessagePlugin.error(e instanceof ApiError ? e.message : '创建失败')
  } finally {
    saving.value = false
  }
}

async function submitEdit() {
  if (!editForm.title.trim()) {
    MessagePlugin.warning('请填写标题')
    return
  }
  saving.value = true
  try {
    await updateTask(editForm.id, {
      title: editForm.title.trim(),
      description: editForm.description.trim(),
      status: editForm.status,
    })
    editVisible.value = false
    MessagePlugin.success('已更新')
    await reload()
  } catch (e) {
    MessagePlugin.error(e instanceof ApiError ? e.message : '更新失败')
  } finally {
    saving.value = false
  }
}

async function removeTask(task: Task) {
  try {
    await deleteTask(task.id)
    MessagePlugin.success('已删除')
    await reload()
  } catch (e) {
    MessagePlugin.error(e instanceof ApiError ? e.message : '删除失败')
  }
}

watch(filterStatus, () => {
  void reload()
})

onMounted(() => {
  void reload()
})
</script>

<template>
  <div class="page">
    <t-card class="shell" :bordered="false">
      <div class="header">
        <div>
          <div class="title">gosilite Tasks</div>
          <div class="subtitle">Vue3 · TDesign · Vite · TypeScript · HTTP API</div>
        </div>
        <t-space>
          <t-button theme="default" variant="outline" @click="reload">
            <template #icon><RefreshIcon /></template>
            刷新
          </t-button>
          <t-button theme="primary" @click="openCreate">
            <template #icon><AddIcon /></template>
            新建任务
          </t-button>
        </t-space>
      </div>

      <t-space class="filters" break-line>
        <t-radio-group v-model="filterStatus" variant="default-filled">
          <t-radio-button
            v-for="opt in STATUS_OPTIONS"
            :key="opt.value"
            :value="opt.value"
          >
            {{ opt.label }}
          </t-radio-button>
        </t-radio-group>
        <span class="count">{{ filteredCount }} 条</span>
      </t-space>

      <t-table
        row-key="id"
        :data="tasks"
        :columns="columns"
        :loading="loading"
        :hover="true"
        :bordered="true"
        size="large"
      >
        <template #status="{ row }">
          <t-tag :theme="STATUS_TAG[row.status as TaskStatus]?.theme ?? 'default'">
            {{ STATUS_TAG[row.status as TaskStatus]?.label ?? row.status }}
          </t-tag>
        </template>
        <template #updated_at="{ row }">
          {{ formatTime(row.updated_at) }}
        </template>
        <template #ops="{ row }">
          <t-space>
            <t-button size="small" variant="text" @click="openEdit(row as Task)">
              <template #icon><EditIcon /></template>
              编辑
            </t-button>
            <t-popconfirm
              content="确认删除该任务？"
              @confirm="removeTask(row as Task)"
            >
              <t-button size="small" variant="text" theme="danger">
                <template #icon><DeleteIcon /></template>
                删除
              </t-button>
            </t-popconfirm>
          </t-space>
        </template>
      </t-table>
    </t-card>

    <t-dialog
      v-model:visible="createVisible"
      header="新建任务"
      :confirm-btn="{ loading: saving, content: '创建' }"
      :on-confirm="submitCreate"
    >
      <t-form label-width="80px">
        <t-form-item label="标题" required>
          <t-input v-model="createForm.title" placeholder="必填" />
        </t-form-item>
        <t-form-item label="描述">
          <t-textarea v-model="createForm.description" :autosize="{ minRows: 3 }" />
        </t-form-item>
        <t-form-item label="状态">
          <t-select v-model="createForm.status">
            <t-option value="todo" label="待办" />
            <t-option value="doing" label="进行中" />
            <t-option value="done" label="完成" />
          </t-select>
        </t-form-item>
      </t-form>
    </t-dialog>

    <t-dialog
      v-model:visible="editVisible"
      header="编辑任务"
      :confirm-btn="{ loading: saving, content: '保存' }"
      :on-confirm="submitEdit"
    >
      <t-form label-width="80px">
        <t-form-item label="标题" required>
          <t-input v-model="editForm.title" placeholder="必填" />
        </t-form-item>
        <t-form-item label="描述">
          <t-textarea v-model="editForm.description" :autosize="{ minRows: 3 }" />
        </t-form-item>
        <t-form-item label="状态">
          <t-select v-model="editForm.status">
            <t-option value="todo" label="待办" />
            <t-option value="doing" label="进行中" />
            <t-option value="done" label="完成" />
          </t-select>
        </t-form-item>
      </t-form>
    </t-dialog>
  </div>
</template>

<style scoped>
.page {
  padding: 24px;
  max-width: 1100px;
  margin: 0 auto;
}

.shell {
  border-radius: 12px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.04);
}

.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 16px;
}

.title {
  font-size: 22px;
  font-weight: 600;
  letter-spacing: 0.2px;
}

.subtitle {
  margin-top: 4px;
  color: rgba(0, 0, 0, 0.55);
  font-size: 13px;
}

.filters {
  margin-bottom: 12px;
  align-items: center;
}

.count {
  color: rgba(0, 0, 0, 0.45);
  font-size: 13px;
}
</style>
