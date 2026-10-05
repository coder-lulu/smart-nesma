/**
 * 通用文件下载工具
 * 支持 xlsx、docx、doc、pdf 等格式的文件下载
 */

/**
 * 下载文件
 * @param {string} fileUrl - 文件URL
 * @param {string} fileName - 文件名
 * @param {string} fileType - 文件类型 (xlsx, docx, doc, pdf, etc.)
 */
export const downloadFile = (fileUrl, fileName, fileType = '') => {
  // 创建下载链接
  const link = document.createElement('a')
  link.href = fileUrl
  link.download = fileName || 'download'
  
  // 设置适当的 MIME 类型
  if (fileType) {
    const mimeTypes = {
      'xlsx': 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
      'docx': 'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
      'doc': 'application/msword',
      'pdf': 'application/pdf',
      'txt': 'text/plain',
      'csv': 'text/csv'
    }
    
    if (mimeTypes[fileType.toLowerCase()]) {
      link.type = mimeTypes[fileType.toLowerCase()]
    }
  }
  
  // 触发下载
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
}

/**
 * 通过 Blob 下载文件
 * @param {Blob} blob - 文件 Blob 对象
 * @param {string} fileName - 文件名
 * @param {string} fileType - 文件类型
 */
export const downloadBlob = (blob, fileName, fileType = '') => {
  // 创建 Blob URL
  const url = window.URL.createObjectURL(blob)
  
  // 创建下载链接
  const link = document.createElement('a')
  link.href = url
  link.download = fileName || 'download'
  
  // 设置适当的 MIME 类型
  if (fileType) {
    const mimeTypes = {
      'xlsx': 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
      'docx': 'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
      'doc': 'application/msword',
      'pdf': 'application/pdf',
      'txt': 'text/plain',
      'csv': 'text/csv'
    }
    
    if (mimeTypes[fileType.toLowerCase()]) {
      link.type = mimeTypes[fileType.toLowerCase()]
    }
  }
  
  // 触发下载
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  
  // 清理 Blob URL
  window.URL.revokeObjectURL(url)
}

/**
 * 通过 API 响应下载文件
 * @param {Response} response - API 响应对象
 * @param {string} fileName - 文件名
 * @param {string} fileType - 文件类型
 */
export const downloadFromResponse = async (response, fileName, fileType = '') => {
  try {
    // 获取文件 Blob
    const blob = await response.blob()
    
    // 从响应头获取文件名（如果服务器提供）
    let finalFileName = fileName
    const contentDisposition = response.headers.get('content-disposition')
    if (contentDisposition) {
      const filenameMatch = contentDisposition.match(/filename[^;=\n]*=((['"]).*?\2|[^;\n]*)/)
      if (filenameMatch && filenameMatch[1]) {
        finalFileName = filenameMatch[1].replace(/['"]/g, '')
      }
    }
    
    // 下载文件
    downloadBlob(blob, finalFileName, fileType)
    
    return true
  } catch (error) {
    console.error('下载文件失败:', error)
    return false
  }
}

/**
 * 通过 Base64 下载文件
 * @param {string} base64Data - Base64 编码的文件数据
 * @param {string} fileName - 文件名
 * @param {string} fileType - 文件类型
 */
export const downloadFromBase64 = (base64Data, fileName, fileType = '') => {
  try {
    // 移除 Base64 前缀（如果存在）
    const base64 = base64Data.replace(/^data:[^;]+;base64,/, '')
    
    // 转换为 Blob
    const byteCharacters = atob(base64)
    const byteNumbers = new Array(byteCharacters.length)
    for (let i = 0; i < byteCharacters.length; i++) {
      byteNumbers[i] = byteCharacters.charCodeAt(i)
    }
    const byteArray = new Uint8Array(byteNumbers)
    
    // 确定 MIME 类型
    let mimeType = 'application/octet-stream'
    if (fileType) {
      const mimeTypes = {
        'xlsx': 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
        'docx': 'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
        'doc': 'application/msword',
        'pdf': 'application/pdf',
        'txt': 'text/plain',
        'csv': 'text/csv'
      }
      
      if (mimeTypes[fileType.toLowerCase()]) {
        mimeType = mimeTypes[fileType.toLowerCase()]
      }
    }
    
    const blob = new Blob([byteArray], { type: mimeType })
    downloadBlob(blob, fileName, fileType)
    
    return true
  } catch (error) {
    console.error('下载 Base64 文件失败:', error)
    return false
  }
}

/**
 * 获取文件扩展名
 * @param {string} fileName - 文件名
 * @returns {string} 文件扩展名
 */
export const getFileExtension = (fileName) => {
  if (!fileName) return ''
  const lastDotIndex = fileName.lastIndexOf('.')
  return lastDotIndex > -1 ? fileName.substring(lastDotIndex + 1).toLowerCase() : ''
}

/**
 * 根据文件类型获取 MIME 类型
 * @param {string} fileType - 文件类型
 * @returns {string} MIME 类型
 */
export const getMimeType = (fileType) => {
  const mimeTypes = {
    'xlsx': 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
    'docx': 'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
    'doc': 'application/msword',
    'pdf': 'application/pdf',
    'txt': 'text/plain',
    'csv': 'text/csv',
    'png': 'image/png',
    'jpg': 'image/jpeg',
    'jpeg': 'image/jpeg',
    'gif': 'image/gif',
    'bmp': 'image/bmp',
    'webp': 'image/webp'
  }
  
  return mimeTypes[fileType.toLowerCase()] || 'application/octet-stream'
}

/**
 * 格式化文件大小
 * @param {number} bytes - 字节数
 * @returns {string} 格式化后的文件大小
 */
export const formatFileSize = (bytes) => {
  if (!bytes || bytes === 0) return '0 B'
  
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
} 