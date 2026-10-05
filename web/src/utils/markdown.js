// 使用 marked 和 highlight.js 的专业 Markdown 渲染器
import { marked } from 'marked'
import hljs from 'highlight.js'

// 配置 marked
marked.setOptions({
  highlight: function(code, lang) {
    // 确保 code 是字符串
    const codeStr = String(code || '')
    
    if (lang && hljs.getLanguage(lang)) {
      try {
        return hljs.highlight(codeStr, { language: lang }).value
      } catch (err) {
        console.warn('代码高亮失败:', err)
        return codeStr
      }
    }
    try {
      return hljs.highlightAuto(codeStr).value
    } catch (err) {
      console.warn('自动代码高亮失败:', err)
      return codeStr
    }
  },
  langPrefix: 'hljs language-',
  breaks: true,
  gfm: true,
  tables: true,
  sanitize: false,
  smartLists: true,
  smartypants: true,
  headerIds: false,
  mangle: false
})

// 自定义渲染器
const renderer = new marked.Renderer()

// 自定义代码块渲染
renderer.code = function(code, infostring, escaped) {
  // 更健壮的语言标识提取
  let lang = ''
  if (infostring) {
    const match = infostring.trim().match(/^(\S+)/)
    lang = match ? match[1] : ''
  }
  
  const langClass = lang ? `language-${lang}` : ''
  const langLabel = lang || 'text'
  
  // 生成唯一ID
  const codeId = `code-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`
  
  // 确保 code 是字符串
  const codeStr = String(code || '')
  
  // 进行代码高亮
  let highlightedCode = codeStr
  if (!escaped) {
    if (lang && hljs.getLanguage(lang)) {
      try {
        highlightedCode = hljs.highlight(codeStr, { language: lang }).value
      } catch (err) {
        console.warn(`代码高亮失败 (${lang}):`, err)
        highlightedCode = codeStr
      }
    } else {
      try {
        const result = hljs.highlightAuto(codeStr)
        highlightedCode = result.value
        // 如果自动检测到语言且没有指定语言，使用检测到的语言
        if (!lang && result.language) {
          lang = result.language
        }
      } catch (err) {
        console.warn('自动代码高亮失败:', err)
        highlightedCode = codeStr
      }
    }
  }
  
  // 最终的语言标签
  const finalLangLabel = lang || 'text'
  
  return `<div class="code-block-container" data-language="${finalLangLabel}" style="margin: 16px 0; border-radius: 8px; overflow: hidden; background: #0f172a; border: 1px solid #334155; display: block; position: relative;">
  <div class="code-block-header" style="display: flex; justify-content: space-between; align-items: center; padding: 10px 16px; background: #1e293b; border-bottom: 1px solid #334155; min-height: 42px;">
    <span class="language-label" style="font-size: 13px; color: #e2e8f0; font-weight: 600; text-transform: uppercase; letter-spacing: 0.5px; font-family: 'Consolas', 'Monaco', 'Courier New', monospace;">${finalLangLabel}</span>
    <button class="copy-code-btn" data-code-id="${codeId}" type="button" style="display: flex; align-items: center; gap: 6px; padding: 8px 12px; background: #475569; color: #f1f5f9; border: 1px solid #64748b; border-radius: 6px; font-size: 12px; cursor: pointer; font-weight: 500;">
      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect>
        <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2 2v1"></path>
      </svg>
      复制
    </button>
  </div>
  <pre class="code-block" id="${codeId}" style="background: #0f172a; color: #f1f5f9; padding: 18px; margin: 0; font-family: 'Fira Code', 'Consolas', 'Monaco', 'Courier New', monospace; font-size: 14px; line-height: 1.6; overflow-x: auto; white-space: pre;"><code class="hljs ${langClass}">${highlightedCode}</code></pre>
</div>`
}

// 自定义行内代码渲染
renderer.codespan = function(code) {
  const codeStr = String(code || '')
  return `<code class="inline-code">${codeStr}</code>`
}

// 自定义表格渲染
renderer.table = function(header, body) {
  const headerStr = String(header || '')
  const bodyStr = String(body || '')
  return `
    <table class="markdown-table">
      <thead>${headerStr}</thead>
      <tbody>${bodyStr}</tbody>
    </table>
  `
}

// 自定义引用渲染
renderer.blockquote = function(quote) {
  const quoteStr = String(quote || '')
  return `<blockquote class="markdown-blockquote">${quoteStr}</blockquote>`
}

// 自定义分割线渲染
renderer.hr = function() {
  return '<hr class="markdown-hr">'
}

// 自定义图片渲染
renderer.image = function(href, title, text) {
  const hrefStr = String(href || '')
  const titleStr = String(title || '')
  const textStr = String(text || '')
  return `<img src="${hrefStr}" alt="${textStr}" title="${titleStr}" class="markdown-image" loading="lazy">`
}

// 自定义链接渲染
renderer.link = function(href, title, text) {
  const hrefStr = String(href || '')
  const titleStr = String(title || '')
  const textStr = String(text || '')
  return `<a href="${hrefStr}" target="_blank" rel="noopener noreferrer" class="markdown-link" title="${titleStr}">${textStr}</a>`
}

// 自定义标题渲染
renderer.heading = function(text, level) {
  const textStr = String(text || '')
  const levelNum = Number(level) || 1
  return `<h${levelNum} class="markdown-header h${levelNum}">${textStr}</h${levelNum}>`
}

// 自定义段落渲染
renderer.paragraph = function(text) {
  const textStr = String(text || '')
  return `<p class="markdown-paragraph">${textStr}</p>`
}

// 自定义列表渲染
renderer.list = function(body, ordered, start) {
  const bodyStr = String(body || '')
  const type = ordered ? 'ol' : 'ul'
  const className = ordered ? 'markdown-list ordered' : 'markdown-list unordered'
  const startatt = (ordered && start !== 1) ? ` start="${start}"` : ''
  return `<${type} class="${className}"${startatt}>${bodyStr}</${type}>`
}

// 自定义列表项渲染
renderer.listitem = function(text) {
  const textStr = String(text || '')
  // 检查是否是任务列表
  if (textStr.includes('<input type="checkbox"')) {
    const isCompleted = textStr.includes('checked')
    return `<li class="task-list-item ${isCompleted ? 'completed' : ''}">${textStr}</li>`
  }
  return `<li>${textStr}</li>`
}

// 自定义强调渲染
renderer.strong = function(text) {
  const textStr = String(text || '')
  return `<strong class="markdown-bold">${textStr}</strong>`
}

// 自定义斜体渲染
renderer.em = function(text) {
  const textStr = String(text || '')
  return `<em class="markdown-italic">${textStr}</em>`
}

// 自定义删除线渲染
renderer.del = function(text) {
  const textStr = String(text || '')
  return `<del class="markdown-strikethrough">${textStr}</del>`
}

// 设置自定义渲染器
marked.use({ renderer })

// 任务列表扩展
const taskListExtension = {
  name: 'taskList',
  level: 'block',
  start(src) {
    const match = src.match(/^(\s*[-*+])\s+\[([ x])\]\s/)
    return match ? match.index : undefined
  },
  tokenizer(src) {
    const rule = /^(\s*)([-*+])\s+\[([ x])\]\s+(.*)$/gm
    const match = rule.exec(src)
    if (match) {
      return {
        type: 'taskList',
        raw: match[0],
        checked: match[3] === 'x',
        text: match[4]
      }
    }
  },
  renderer(token) {
    const checked = token.checked ? 'checked' : ''
    const completed = token.checked ? 'completed' : ''
    const textStr = String(token.text || '')
    return `<li class="task-list-item ${completed}">
      <input type="checkbox" ${checked} disabled> 
      ${this.parser.parseInline(textStr)}
    </li>`
  }
}

// 注册任务列表扩展
marked.use({ extensions: [taskListExtension] })

// 数学公式扩展（简单版本）
const mathExtension = {
  name: 'math',
  level: 'inline',
  start(src) {
    const match = src.match(/\$/)
    return match ? match.index : undefined
  },
  tokenizer(src) {
    const rule = /^\$([^$\n]+)\$/
    const match = rule.exec(src)
    if (match) {
      return {
        type: 'math',
        raw: match[0],
        text: match[1]
      }
    }
  },
  renderer(token) {
    const textStr = String(token.text || '')
    return `<span class="math-inline">${textStr}</span>`
  }
}

// 块级数学公式扩展
const mathBlockExtension = {
  name: 'mathBlock',
  level: 'block',
  start(src) {
    const match = src.match(/\$\$/)
    return match ? match.index : undefined
  },
  tokenizer(src) {
    const rule = /^\$\$([\s\S]*?)\$\$/
    const match = rule.exec(src)
    if (match) {
      return {
        type: 'mathBlock',
        raw: match[0],
        text: match[1]
      }
    }
  },
  renderer(token) {
    const textStr = String(token.text || '')
    return `<div class="math-block">${textStr}</div>`
  }
}

// 注册数学公式扩展
marked.use({ extensions: [mathExtension, mathBlockExtension] })

class MarkdownRenderer {
  constructor() {
    this.streamingContent = ''
  }

  // 渲染完整的 Markdown 文本
  render(text) {
    if (!text || typeof text !== 'string') {
      return ''
    }

    try {
      // 预处理任务列表
      const processedText = this.preprocessTaskLists(text)
      
      // 使用 marked 渲染
      const html = marked.parse(processedText)
      
      // 后处理，包装任务列表
      const finalHtml = this.postprocessTaskLists(html)
      
      return finalHtml
    } catch (error) {
      console.error('Markdown 渲染失败:', error)
      // 回退到基本的文本处理
      return text.replace(/\n/g, '<br>')
    }
  }

  // 预处理任务列表
  preprocessTaskLists(text) {
    return text.replace(/^(\s*[-*+])\s+\[([ x])\]\s+(.*)$/gm, (match, bullet, check, content) => {
      const checked = check === 'x'
      return `${bullet} __TASK_${checked ? 'CHECKED' : 'UNCHECKED'}__${content}`
    })
  }

  // 后处理任务列表
  postprocessTaskLists(html) {
    // 包装任务列表项
    html = html.replace(/__TASK_(CHECKED|UNCHECKED)__/g, (match, status) => {
      const checked = status === 'CHECKED'
      return `<input type="checkbox" ${checked ? 'checked' : ''} disabled> `
    })

    // 将包含任务列表项的 ul 添加特殊类名
    html = html.replace(/<ul([^>]*)>([\s\S]*?<input type="checkbox"[\s\S]*?)<\/ul>/g, (match, attrs, content) => {
      return `<ul class="markdown-task-list"${attrs}>${content}</ul>`
    })

    return html
  }

  // 流式渲染（用于打字机效果）
  renderStreaming(text, isComplete = false) {
    if (!text) return ''
    
    let html = this.render(text)
    
    // 如果还在流式输出中，添加光标
    if (!isComplete) {
      html += '<span class="typing-cursor">|</span>'
    }
    
    return html
  }

  // 清理HTML（防止XSS）
  sanitizeHtml(html) {
    const div = document.createElement('div')
    div.textContent = html
    return div.innerHTML
  }
}

// 复制代码功能已移至Vue组件中处理，通过事件委托实现

// 创建全局实例
export const markdownRenderer = new MarkdownRenderer()

export default MarkdownRenderer