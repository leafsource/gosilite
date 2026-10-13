<template>
  <div class="upload-file">
    <t-upload
      v-model="fileList"
      multiple
      :abridge-name="abridgeName"
      :theme="theme"
      :accept="rawAccept"
      :action="resolvedAction"
      :before-upload="handleBeforeUpload"
      :max="limit"
      :headers="headers"
      :draggable="draggable"
      :size-limit="{ size: fileSize, unit: 'MB', message: '上传文件大小不能超过 {sizeLimit} MB!' }"
      :disabled="disabled"
      :allow-upload-duplicate-file="allowUploadDuplicateFile"
      @one-file-success="handleOneUploadSuccess"
      @one-file-fail="handleUploadError"
      @success="handleUploadSuccess"
      @remove="handleDelete"
      @validate="onValidate"
    >
      <template v-if="isShowTip" #tips>
        请上传大小不超过 {{ fileSize }}MB 的文件
        <t-tooltip>
          <template #content>
            <p v-for="item in rawAccept?.split(',')" :key="item" style="word-break: break-all">
              {{ item }}
            </p>
          </template>
          <t-link theme="primary">查看格式要求</t-link>
        </t-tooltip>
      </template>
      <!-- 上传按钮 -->
      <slot />
      <t-button v-if="!$slots.default" variant="outline">
        <template #icon>
          <cloud-upload-icon />
        </template>
        选取文件
      </t-button>
    </t-upload>
  </div>
</template>
<script lang="ts" setup>
import { CloudUploadIcon } from 'tdesign-icons-vue-next';
import type { SuccessContext, UploadFile, UploadRemoveContext, UploadValidateType } from 'tdesign-vue-next';
import { MessagePlugin } from 'tdesign-vue-next';
import { computed, ref, watch } from 'vue';

import { delOss, listByIds, listByUrls } from '@/api/system/oss';
import { useUserStore } from '@/store';
import { getHttpFileName, getHttpFileSuffix } from '@/utils/file';

defineOptions({ name: 'SUploadFile' });

const props = withDefaults(defineProps<FileUploadProps>(), {
  limit: 5,
  fileSize: 5,
  fileType: () => ['doc', 'docx', 'xlsx', 'xls', 'csv', 'ppt', 'pptx', 'txt', 'pdf'],
  isShowTip: true,
  theme: 'file',
  mode: 'id',
  disabled: false,
  allowUploadDuplicateFile: false,
  showSuccessMessage: false,
  action: '/common/file/upload',
  deleteRemoteOnRemove: false,
});

const emit = defineEmits(['update:model-value', 'change', 'success']);

export interface FileUploadProps {
  modelValue?: string | string[];
  // 文件名过长时，需要省略中间的文本，保留首尾文本。示例：[10, 7]，表示首尾分别保留的文本长度。
  abridgeName?: Array<number>;
  // 文件数量限制,为0不限制
  limit?: number;
  // 是否支持拖拽上传
  draggable?: boolean;
  // 大小限制(MB)
  fileSize?: number;
  // 接受上传的文件类型
  accept?: Array<string>;
  // 文件类型, 例如['doc', 'docx']
  fileType?: string[];
  // 是否显示提示
  isShowTip?: boolean;
  theme?: 'custom' | 'file' | 'file-input' | 'file-flow';
  // 模式，id模式返回文件 id，url模式返回url链接
  mode?: 'id' | 'url';
  // 禁用组件
  disabled?: boolean;
  // 是否允许重复上传相同文件名的文件
  allowUploadDuplicateFile?: boolean;
  // 上传成功后是否弹「文件【x】上传成功！」通用提示（默认关闭，由业务页自行决定成功提示，避免双重提醒）
  showSuccessMessage?: boolean;
  // 上传端点（gosilite：内部拼 VITE_API_URL_PREFIX 前缀）
  action?: string;
  // 移除文件时是否同时删除台账记录（对应 ruoyi supportSelectFile=false 的直传清理语义；
  // 默认 false = gosilite 附件表单语义「仅清本地，远程由业务删除接口决定」）
  deleteRemoteOnRemove?: boolean;
}

/** 上传成功负载（调用方需要 id/名称） */
export interface SUploadSuccessFile {
  ossId?: string;
  url: string;
  name: string;
  size?: number;
}

const userStore = useUserStore();

/** t-upload 走原生 XHR 不经 axios：以 / 开头的相对路径需拼 VITE_API_URL_PREFIX（vite proxy 再剥离前缀转发后端） */
const resolvedAction = computed(() => {
  const prefix = (import.meta.env.VITE_API_URL_PREFIX || '').replace(/\/$/, '');
  const action = props.action;
  if (!prefix || !action.startsWith('/')) return action;
  return `${prefix}${action}`;
});

/** computed 而非一次性 ref：beforeUpload 里刷新 token 后新 token 立即生效 */
const headers = computed(() => ({ Authorization: `Bearer ${userStore.token}` }));

const rawAccept = computed(() => {
  return props.accept?.join(',') || props.fileType.map((value) => `.${value}`).join(',');
});

const fileList = ref<UploadFile[]>([]);

/** 会话内上传元数据缓存：url/ossId → 元数据，modelValue 回写重建列表时保留 originalName/url */
const metaCache = new Map<string, { name?: string; size?: number; url?: string; ossId?: string }>();

function cacheMeta(items: SUploadSuccessFile[]) {
  items.forEach((item) => {
    const meta = { name: item.name, size: item.size, url: item.url, ossId: item.ossId };
    if (item.url) metaCache.set(item.url, meta);
    if (item.ossId) metaCache.set(item.ossId, meta);
  });
}

const splitHttp = /,(?=http)/;
watch(
  () => props.modelValue,
  async (val) => {
    if (val) {
      let temp = 1;
      // 首先将值转为数组
      let list: any[];
      if (Array.isArray(val)) {
        list = val;
      } else if (props.mode === 'url') {
        // gosilite 回显：会话缓存 → 台账按 key 查询（无权限返回空）→ URL 文件名兜底
        const urls = val.split(splitHttp);
        const tempMap = await listByUrls(urls);
        list = urls.map((url: string) => {
          const cached = metaCache.get(url);
          const oss = tempMap.get(url);
          return (
            (oss && {
              name: oss.originalName,
              status: 'success',
              size: oss.size,
              uploadTime: oss.createdAt,
              url: oss.url,
              ossId: oss.id,
            }) ||
            (cached && {
              name: cached.name,
              status: 'success',
              size: cached.size,
              url,
              ossId: cached.ossId,
            }) || { name: getHttpFileName(url), status: 'success', size: 0, url }
          );
        });
      } else {
        // mode='id'：会话缓存 → 逐个查台账 → id 兜底
        const tempMap = await listByIds(val);
        list = String(val)
          .split(',')
          .filter(Boolean)
          .map((id) => {
            const cached = metaCache.get(id);
            const oss = tempMap.get(id);
            return (
              (oss && {
                name: oss.originalName,
                status: 'success',
                size: oss.size,
                uploadTime: oss.createdAt,
                url: oss.url,
                ossId: oss.id,
              }) ||
              (cached && {
                name: cached.name,
                status: 'success',
                size: cached.size,
                url: cached.url,
                ossId: cached.ossId,
              }) || { name: id, status: 'success', size: 0, url: '' }
            );
          });
      }
      // 值已在异步等待期间变化：丢弃过期回显，避免覆盖新状态
      if (props.modelValue !== val) return;
      // 然后将数组转为对象数组
      fileList.value = list.map((item: any) => {
        if (typeof item === 'string') {
          item = { name: getHttpFileName(item), status: 'success', url: item };
        } else {
          item = {
            name: item.name,
            status: item.status,
            size: item.size,
            uploadTime: item.uploadTime,
            url: item.url,
            ossId: item.ossId,
          };
        }
        item.uid = item.uid || Date.now() + temp++;
        return item;
      });
    } else {
      fileList.value = [];
    }
  },
  { deep: true, immediate: true },
);

// 有文件数量超出时会触发，文件大小超出限制、文件同名时会触发等场景。注意如果设置允许上传同名文件，则此事件不会触发
const onValidate = (params: { type: UploadValidateType; files: UploadFile[] }) => {
  const { files, type } = params;
  if (type === 'FILE_OVER_SIZE_LIMIT') {
    MessagePlugin.warning(`${files.map((t) => t.name).join('、')} 等文件大小超出限制，已自动过滤`, 5000);
  } else if (type === 'FILES_OVER_LENGTH_LIMIT') {
    MessagePlugin.warning('文件数量超出限制，仅上传未超出数量的文件');
  } else if (type === 'FILTER_FILE_SAME_NAME') {
    // 如果希望支持上传同名文件，请设置 allowUploadDuplicateFile={true}
    MessagePlugin.warning('不允许上传同名文件');
  }
};

// 上传前校检格式；先确保 token 新鲜（t-upload 原生 XHR 绕过 axios 401→refresh，token 过期会直接 401）
async function handleBeforeUpload(file: UploadFile) {
  try {
    await userStore.ensureFreshAccessToken();
  } catch {
    // refresh 失败时 store 已处理跳登录，中止上传
    return false;
  }
  const fileName = file.name ?? '';
  // 校检文件类型
  if (props.fileType?.length) {
    const fileExt = getHttpFileSuffix(fileName);
    const isTypeOk = props.fileType.includes(fileExt);
    if (!isTypeOk) {
      MessagePlugin.error(`文件格式不正确, 请上传${props.fileType.join('/')}格式文件!`);
      return false;
    }
  }
  // 校检文件名是否包含特殊字符
  if (fileName.includes(',')) {
    MessagePlugin.error('文件名不正确，不能包含英文逗号!');
    return false;
  }
  return true;
}

/**
 * 单文件上传结束回调（multiple 且一文件一请求时逐个触发）。
 * gosilite 信封 {code,msg,data:{id,url,originalName,size}}：code!==200 时 HTTP 仍为 200，t-upload 视为成功，此处判错提示。
 */
function handleOneUploadSuccess(context: Pick<SuccessContext, 'e' | 'file' | 'response' | 'XMLHttpRequest'>) {
  const response = context.response as { code?: number; msg?: string; data?: { originalName?: string } } | null;
  if (response?.code !== 200) {
    MessagePlugin.error(response?.msg || '上传失败');
  } else if (props.showSuccessMessage) {
    MessagePlugin.success(`文件【${response.data?.originalName ?? context.file?.name}】上传成功！`);
  }
}

// 单文件失败回调（HTTP 层失败：响应体可能有业务 msg）
function handleUploadError(options: { response?: { msg?: string } | null; XMLHttpRequest?: XMLHttpRequest }) {
  let msg = options?.response?.msg;
  if (!msg && options?.XMLHttpRequest?.responseText) {
    try {
      msg = (JSON.parse(options.XMLHttpRequest.responseText) as { msg?: string }).msg;
    } catch {
      // 非 JSON 响应（如网关 HTML 页）走兜底文案
    }
  }
  MessagePlugin.error(msg || '上传文件失败');
}

// 上传成功回调（multiple 多请求：results 为逐文件结果，含成功与失败）
function handleUploadSuccess(context: SuccessContext) {
  const uploadList = (context.results ?? [])
    .filter((value) => value.response?.code === 200 && value.response?.data?.url)
    .map<SUploadSuccessFile>((value) => ({
      name: value.response?.data?.originalName ?? '',
      url: value.response?.data?.url ?? '',
      ossId: value.response?.data?.id,
      size: value.response?.data?.size,
    }));
  cacheMeta(uploadList);
  uploadedSuccessfully(uploadList);
  emit('success', uploadList);
}

// 删除文件
function handleDelete({ file }: UploadRemoveContext) {
  const { ossId, name } = file as UploadFile & { ossId?: string };
  // 直传清理（ruoyi supportSelectFile=false 语义）：默认关闭，开启且无权限时静默跳过
  if (props.deleteRemoteOnRemove && ossId) {
    delOss(String(ossId))
      .then(() => {
        MessagePlugin.success(`文件【${name}】删除成功！`);
      })
      .catch(() => undefined);
  }
  const value = listToString(fileList.value);
  emit('update:model-value', value);
  emit('change', value);
}

/**
 * 上传结束处理。t-upload 成功后已把文件写入 fileList（但 gosilite url 在 data 内 →
 * 新文件 url 为 undefined）：先滤掉未带 url 的本批新文件，再并入从信封映射出的 uploadList，避免重复条目。
 */
function uploadedSuccessfully(uploadList: SUploadSuccessFile[]) {
  fileList.value = fileList.value.filter((f) => f.url !== undefined).concat(uploadList);
  const value = listToString(fileList.value);
  emit('update:model-value', value);
  emit('change', value);
}

// 对象转成指定字符串分隔
function listToString(list: UploadFile[], separator?: string) {
  let strs = '';
  separator = separator || ',';
  for (const i in list) {
    if (props.mode === 'url') {
      if (list[i].url) {
        strs += list[i].url + separator;
      }
    } else if (props.mode === 'id') {
      if (list[i].ossId) {
        strs += list[i].ossId + separator;
      }
    }
  }
  return strs !== '' ? strs.substring(0, strs.length - 1) : '';
}
</script>
