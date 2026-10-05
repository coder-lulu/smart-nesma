/**
 * 下载工具使用示例
 * 展示如何使用各种下载功能
 */

import { 
  downloadFile, 
  downloadBlob, 
  downloadFromResponse, 
  downloadFromBase64,
  getFileExtension,
  getMimeType,
  formatFileSize 
} from './downloadFile'

import {
  downloadExcel,
  downloadWord,
  downloadWordDoc,
  downloadPDF,
  downloadByType,
  downloadNesmaDocument,
  batchDownloadDocuments
} from './officeDownload'

/**
 * 示例：下载 Excel 文件
 */
export const exampleDownloadExcel = async () => {
  try {
    // 假设从 API 获取 Excel 文件
    const response = await fetch('/api/excel/download')
    const blob = await response.blob()
    
    // 方法1：使用通用下载工具
    downloadBlob(blob, 'report.xlsx', 'xlsx')
    
    // 方法2：使用专门的 Excel 下载工具
    downloadExcel(blob, 'report', { 
      includeTimestamp: true,
      suffix: 'v1.0'
    })
    
    // 方法3：使用类型自动下载
    downloadByType(blob, 'report', 'excel', {
      includeTimestamp: true
    })
  } catch (error) {
    console.error('下载 Excel 失败:', error)
  }
}

/**
 * 示例：下载 Word 文档
 */
export const exampleDownloadWord = async () => {
  try {
    // 假设从 API 获取 Word 文档
    const response = await fetch('/api/word/download')
    const blob = await response.blob()
    
    // 方法1：下载 docx 格式
    downloadWord(blob, 'document', {
      includeTimestamp: false,
      suffix: 'final'
    })
    
    // 方法2：下载 doc 格式（旧版）
    downloadWordDoc(blob, 'document', {
      includeTimestamp: true
    })
    
    // 方法3：使用类型自动下载
    downloadByType(blob, 'document', 'word', {
      includeTimestamp: true
    })
  } catch (error) {
    console.error('下载 Word 失败:', error)
  }
}

/**
 * 示例：下载 PDF 文档
 */
export const exampleDownloadPDF = async () => {
  try {
    // 假设从 API 获取 PDF 文档
    const response = await fetch('/api/pdf/download')
    const blob = await response.blob()
    
    // 下载 PDF
    downloadPDF(blob, 'report', {
      includeTimestamp: true,
      suffix: 'final'
    })
    
    // 或者使用类型自动下载
    downloadByType(blob, 'report', 'pdf', {
      includeTimestamp: true
    })
  } catch (error) {
    console.error('下载 PDF 失败:', error)
  }
}

/**
 * 示例：下载 NESMA 文档
 */
export const exampleDownloadNesmaDocument = async (documentId) => {
  try {
    // 假设从 API 获取 NESMA 文档
    const response = await fetch(`/api/nesma/document/download?documentId=${documentId}`)
    const blob = await response.blob()
    
    // 文档信息
    const documentInfo = {
      name: '需求规格说明书',
      version: 'v1.0.0',
      type: 'word',
      format: 'requirement_spec'
    }
    
    // 使用专门的 NESMA 文档下载工具
    downloadNesmaDocument(blob, documentInfo)
  } catch (error) {
    console.error('下载 NESMA 文档失败:', error)
  }
}

/**
 * 示例：批量下载文档
 */
export const exampleBatchDownload = async (documentIds) => {
  try {
    const documents = []
    
    // 准备下载函数
    const downloadFunc = async (document) => {
      const response = await fetch(`/api/nesma/document/download?documentId=${document.id}`)
      const blob = await response.blob()
      downloadNesmaDocument(blob, document)
    }
    
    // 批量下载
    await batchDownloadDocuments(documents, downloadFunc, {
      delay: 2000, // 2秒延迟
      onProgress: (current, total, document) => {
        console.log(`下载进度: ${current}/${total} - ${document.name}`)
      }
    })
  } catch (error) {
    console.error('批量下载失败:', error)
  }
}

/**
 * 示例：从 Base64 下载文件
 */
export const exampleDownloadFromBase64 = () => {
  // 假设有 Base64 编码的文件数据
  const base64Data = 'data:application/vnd.openxmlformats-officedocument.spreadsheetml.sheet;base64,UEsDBBQAAAAIAA...'
  
  // 下载 Excel 文件
  downloadFromBase64(base64Data, 'report.xlsx', 'xlsx')
  
  // 下载 Word 文档
  const wordBase64 = 'data:application/vnd.openxmlformats-officedocument.wordprocessingml.document;base64,UEsDBBQAAAAIAA...'
  downloadFromBase64(wordBase64, 'document.docx', 'docx')
}

/**
 * 示例：工具函数使用
 */
export const exampleUtilityFunctions = () => {
  // 获取文件扩展名
  const ext = getFileExtension('document.xlsx') // 返回 'xlsx'
  
  // 获取 MIME 类型
  const mimeType = getMimeType('xlsx') // 返回 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet'
  
  // 格式化文件大小
  const size = formatFileSize(1024 * 1024) // 返回 '1 MB'
  
  console.log('文件扩展名:', ext)
  console.log('MIME 类型:', mimeType)
  console.log('文件大小:', size)
}

/**
 * 示例：处理不同格式的响应
 */
export const exampleHandleDifferentResponses = async () => {
  try {
    // 1. 处理 Blob 响应
    const blobResponse = await fetch('/api/file/blob')
    const blob = await blobResponse.blob()
    downloadBlob(blob, 'file.xlsx', 'xlsx')
    
    // 2. 处理 Response 对象
    const response = await fetch('/api/file/download')
    downloadFromResponse(response, 'file.docx', 'docx')
    
    // 3. 处理 Base64 数据
    const base64Response = await fetch('/api/file/base64')
    const { data } = await base64Response.json()
    downloadFromBase64(data, 'file.pdf', 'pdf')
    
  } catch (error) {
    console.error('处理响应失败:', error)
  }
} 