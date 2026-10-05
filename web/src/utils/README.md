# 文件下载工具使用说明

本项目提供了完整的文件下载解决方案，支持多种文件格式的下载，特别是 Office 文档（xlsx、docx、doc）和 PDF 文档。

## 文件结构

```
utils/
├── downloadFile.js      # 通用文件下载工具
├── officeDownload.js    # Office 文档专用下载工具
├── downloadImg.js       # 图片下载工具（原有）
└── downloadExample.js   # 使用示例
```

## 通用下载工具 (downloadFile.js)

### 主要功能

- `downloadFile(fileUrl, fileName, fileType)` - 通过 URL 下载文件
- `downloadBlob(blob, fileName, fileType)` - 通过 Blob 下载文件
- `downloadFromResponse(response, fileName, fileType)` - 通过 API 响应下载文件
- `downloadFromBase64(base64Data, fileName, fileType)` - 通过 Base64 数据下载文件

### 工具函数

- `getFileExtension(fileName)` - 获取文件扩展名
- `getMimeType(fileType)` - 获取 MIME 类型
- `formatFileSize(bytes)` - 格式化文件大小

## Office 文档下载工具 (officeDownload.js)

### 专门针对 Office 文档的下载函数

- `downloadExcel(data, fileName, options)` - 下载 Excel 文件 (.xlsx)
- `downloadWord(data, fileName, options)` - 下载 Word 文档 (.docx)
- `downloadWordDoc(data, fileName, options)` - 下载旧版 Word 文档 (.doc)
- `downloadPDF(data, fileName, options)` - 下载 PDF 文档
- `downloadByType(data, fileName, documentType, options)` - 根据类型自动下载
- `downloadNesmaDocument(data, documentInfo)` - 下载 NESMA 文档
- `batchDownloadDocuments(documents, downloadFunc, options)` - 批量下载文档

## 使用示例

### 1. 下载 Excel 文件

```javascript
import { downloadExcel } from '@/utils/officeDownload'

// 从 API 获取 Excel 文件
const response = await fetch('/api/excel/download')
const blob = await response.blob()

// 下载 Excel 文件
downloadExcel(blob, 'report', {
  includeTimestamp: true,
  suffix: 'v1.0'
})
// 生成文件名: report_v1.0_1234567890.xlsx
```

### 2. 下载 Word 文档

```javascript
import { downloadWord } from '@/utils/officeDownload'

// 下载 Word 文档
downloadWord(blob, 'document', {
  includeTimestamp: false,
  suffix: 'final'
})
// 生成文件名: document_final.docx
```

### 3. 下载 NESMA 文档

```javascript
import { downloadNesmaDocument } from '@/utils/officeDownload'

// 文档信息
const documentInfo = {
  name: '需求规格说明书',
  version: 'v1.0.0',
  type: 'word',
  format: 'requirement_spec'
}

// 下载 NESMA 文档
downloadNesmaDocument(blob, documentInfo)
// 生成文件名: 需求规格说明书_v1.0.0.docx
```

### 4. 批量下载

```javascript
import { batchDownloadDocuments } from '@/utils/officeDownload'

// 批量下载文档
await batchDownloadDocuments(documents, downloadFunc, {
  delay: 2000, // 2秒延迟
  onProgress: (current, total, document) => {
    console.log(`下载进度: ${current}/${total} - ${document.name}`)
  }
})
```

### 5. 使用通用下载工具

```javascript
import { downloadBlob, downloadFromResponse } from '@/utils/downloadFile'

// 通过 Blob 下载
downloadBlob(blob, 'file.xlsx', 'xlsx')

// 通过 API 响应下载
const response = await fetch('/api/file/download')
downloadFromResponse(response, 'file.docx', 'docx')
```

## 选项参数

### 下载选项 (options)

```javascript
{
  includeTimestamp: false,  // 是否在文件名中包含时间戳
  suffix: ''               // 文件名后缀
}
```

### 批量下载选项

```javascript
{
  delay: 1000,            // 下载间隔延迟（毫秒）
  onProgress: (current, total, document) => {
    // 进度回调函数
  }
}
```

## 支持的文件类型

### Office 文档
- `.xlsx` - Excel 文件
- `.docx` - Word 文档（新版）
- `.doc` - Word 文档（旧版）
- `.pdf` - PDF 文档

### 其他文件
- `.txt` - 文本文件
- `.csv` - CSV 文件
- `.png`, `.jpg`, `.jpeg`, `.gif`, `.bmp`, `.webp` - 图片文件

## MIME 类型映射

| 文件类型 | MIME 类型 |
|---------|-----------|
| xlsx | application/vnd.openxmlformats-officedocument.spreadsheetml.sheet |
| docx | application/vnd.openxmlformats-officedocument.wordprocessingml.document |
| doc | application/msword |
| pdf | application/pdf |
| txt | text/plain |
| csv | text/csv |

## 注意事项

1. **浏览器兼容性**: 确保浏览器支持 Blob 和 URL.createObjectURL
2. **文件大小**: 大文件下载时注意内存使用
3. **批量下载**: 避免同时下载过多文件，建议添加延迟
4. **错误处理**: 始终添加适当的错误处理
5. **清理资源**: 下载完成后及时清理 Blob URL

## 在项目中的使用

### 文档管理页面

```javascript
// 在 DocumentForm.vue 中使用
import { downloadNesmaDocument } from '@/utils/officeDownload'

const handleDownload = async (document) => {
  try {
    const res = await downloadDocument({ documentId: document.ID })
    downloadNesmaDocument(
      new Blob([res], { type: 'application/octet-stream' }),
      document
    )
    ElMessage.success('下载成功')
  } catch (error) {
    ElMessage.error('下载失败')
  }
}
```

### 批量操作

```javascript
// 批量下载选中的文档
const handleBatchDownload = async () => {
  const selectedDocuments = tableData.value.filter(item => selectedIds.value.includes(item.ID))
  
  await batchDownloadDocuments(selectedDocuments, async (document) => {
    const res = await downloadDocument({ documentId: document.ID })
    downloadNesmaDocument(
      new Blob([res], { type: 'application/octet-stream' }),
      document
    )
  }, {
    delay: 2000,
    onProgress: (current, total) => {
      ElMessage.info(`正在下载第 ${current}/${total} 个文件`)
    }
  })
}
```

## 扩展功能

如需添加新的文件类型支持，可以在 `downloadFile.js` 中的 `mimeTypes` 对象中添加相应的 MIME 类型映射。 