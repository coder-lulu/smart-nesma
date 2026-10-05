/**
 * 网站配置文件
 */
const greenText = (text) => `\x1b[32m${text}\x1b[0m`

const config = {
  appName: 'NESMA智能分析系统',
  appLogo: 'logo.png',
  showViteLogo: true,
  logs: []
}

export const viteLogo = (env) => {
  if (config.showViteLogo) {
    console.log(
      greenText(
        `> 欢迎使用NESMA智能分析系统`
      )
    )
  }
}

export default config
