/**
 * Office 文档下载工具
 * 专门用于下载 xlsx、docx、doc 等 Office 文档
 */

import { downloadBlob, downloadFromResponse, getFileExtension } from './downloadFile'

/**
 * 下载 Excel 文件 (xlsx)
 * @param {Blob|Response} data - 文件数据 (Blob 或 Response)
 * @param {string} fileName - 文件名
 * @param {object} options - 选项
 * @param {boolean} options.includeTimestamp - 是否在文件名中包含时间戳
 * @param {string} options.suffix - 文件名后缀
 */
export const downloadExcel = (data, fileName, options = {}) => {
  const { includeTimestamp = false, suffix = '' } = options
  
  let finalFileName = fileName || 'excel_document'
  if (suffix) {
    finalFileName += `_${suffix}`
  }
  if (includeTimestamp) {
    finalFileName += `_${new Date().getTime()}`
  }
  finalFileName += '.xlsx'
  
  if (data instanceof Blob) {
    downloadBlob(data, finalFileName, 'xlsx')
  } else if (data instanceof Response) {
    downloadFromResponse(data, finalFileName, 'xlsx')
  } else {
    console.error('不支持的数据类型')
  }
}

/**
 * 下载 Word 文档 (docx)
 * @param {Blob|Response} data - 文件数据 (Blob 或 Response)
 * @param {string} fileName - 文件名
 * @param {object} options - 选项
 * @param {boolean} options.includeTimestamp - 是否在文件名中包含时间戳
 * @param {string} options.suffix - 文件名后缀
 */
export const downloadWord = (data, fileName, options = {}) => {
  const { includeTimestamp = false, suffix = '' } = options
  
  let finalFileName = fileName || 'word_document'
  if (suffix) {
    finalFileName += `_${suffix}`
  }
  if (includeTimestamp) {
    finalFileName += `_${new Date().getTime()}`
  }
  finalFileName += '.docx'
  
  if (data instanceof Blob) {
    downloadBlob(data, finalFileName, 'docx')
  } else if (data instanceof Response) {
    downloadFromResponse(data, finalFileName, 'docx')
  } else {
    console.error('不支持的数据类型')
  }
}

/**
 * 下载旧版 Word 文档 (doc)
 * @param {Blob|Response} data - 文件数据 (Blob 或 Response)
 * @param {string} fileName - 文件名
 * @param {object} options - 选项
 * @param {boolean} options.includeTimestamp - 是否在文件名中包含时间戳
 * @param {string} options.suffix - 文件名后缀
 */
export const downloadWordDoc = (data, fileName, options = {}) => {
  const { includeTimestamp = false, suffix = '' } = options
  
  let finalFileName = fileName || 'word_document'
  if (suffix) {
    finalFileName += `_${suffix}`
  }
  if (includeTimestamp) {
    finalFileName += `_${new Date().getTime()}`
  }
  finalFileName += '.doc'
  
  if (data instanceof Blob) {
    downloadBlob(data, finalFileName, 'doc')
  } else if (data instanceof Response) {
    downloadFromResponse(data, finalFileName, 'doc')
  } else {
    console.error('不支持的数据类型')
  }
}

/**
 * 下载 PDF 文档
 * @param {Blob|Response} data - 文件数据 (Blob 或 Response)
 * @param {string} fileName - 文件名
 * @param {object} options - 选项
 * @param {boolean} options.includeTimestamp - 是否在文件名中包含时间戳
 * @param {string} options.suffix - 文件名后缀
 */
export const downloadPDF = (data, fileName, options = {}) => {
  const { includeTimestamp = false, suffix = '' } = options
  
  let finalFileName = fileName || 'document'
  if (suffix) {
    finalFileName += `_${suffix}`
  }
  if (includeTimestamp) {
    finalFileName += `_${new Date().getTime()}`
  }
  finalFileName += '.pdf'
  
  if (data instanceof Blob) {
    downloadBlob(data, finalFileName, 'pdf')
  } else if (data instanceof Response) {
    downloadFromResponse(data, finalFileName, 'pdf')
  } else {
    console.error('不支持的数据类型')
  }
}

/**
 * 根据文档类型自动下载
 * @param {Blob|Response} data - 文件数据
 * @param {string} fileName - 文件名
 * @param {string} documentType - 文档类型 (excel, word, pdf)
 * @param {object} options - 选项
 */
export const downloadByType = (data, fileName, documentType, options = {}) => {
  const typeMap = {
    'excel': downloadExcel,
    'word': downloadWord,
    'doc': downloadWordDoc,
    'pdf': downloadPDF
  }
  
  const downloadFunc = typeMap[documentType.toLowerCase()]
  if (downloadFunc) {
    downloadFunc(data, fileName, options)
  } else {
    console.error(`不支持的文档类型: ${documentType}`)
  }
}

/**
 * 下载 NESMA 文档
 * @param {Blob|Response} data - 文件数据
 * @param {object} documentInfo - 文档信息
 * @param {string} documentInfo.name - 文档名称
 * @param {string} documentInfo.version - 文档版本
 * @param {string} documentInfo.type - 文档类型
 * @param {string} documentInfo.format - 文档格式
 */
export const downloadNesmaDocument = (data, documentInfo) => {
  const { name, version, type, format } = documentInfo
  
  // 根据文档类型和格式确定文件扩展名
  let fileExt = 'docx'
  if (type === 'excel' || format === 'business_summary') {
    fileExt = 'xlsx'
  } else if (type === 'word') {
    if (format === 'requirement_spec' || format === 'nesma_report') {
      fileExt = 'docx'
    } else {
      fileExt = 'doc'
    }
  } else if (type === 'pdf') {
    fileExt = 'pdf'
  }
  
  const fileName = `${name}_${version}.${fileExt}`
  
  downloadByType(data, fileName, type, {
    includeTimestamp: false
  })
}

/**
 * 批量下载文档
 * @param {Array} documents - 文档列表
 * @param {Function} downloadFunc - 下载函数
 * @param {object} options - 选项
 */
export const batchDownloadDocuments = async (documents, downloadFunc, options = {}) => {
  const { delay = 1000, onProgress } = options
  
  for (let i = 0; i < documents.length; i++) {
    const document = documents[i]
    
    try {
      await downloadFunc(document)
      
      if (onProgress) {
        onProgress(i + 1, documents.length, document)
      }
      
      // 添加延迟避免浏览器阻止多个下载
      if (i < documents.length - 1) {
        await new Promise(resolve => setTimeout(resolve, delay))
      }
    } catch (error) {
      console.error(`下载文档失败: ${document.name}`, error)
    }
  }
} 