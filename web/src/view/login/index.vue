<template>
  <div id="userLayout" class="relative w-full h-full">
    <div
      class="flex justify-center items-center w-full h-full bg-gradient-to-br from-gray-50 via-blue-50 to-indigo-100 md:w-screen md:h-screen"
    >
              <div class="relative px-6 mx-auto w-full max-w-md">
          <!-- 装饰性背景元素 -->
          <div class="absolute -top-4 -left-4 w-24 h-24 bg-gradient-to-br from-blue-200 to-indigo-300 rounded-2xl opacity-20 transform rotate-12"></div>
          <div class="absolute -right-4 -bottom-4 w-32 h-32 bg-gradient-to-br from-emerald-200 to-blue-300 rounded-full opacity-15"></div>
          <div class="absolute -left-8 top-1/2 w-16 h-16 bg-gradient-to-br from-purple-200 to-pink-300 rounded-xl opacity-25 transform -rotate-45"></div>
          
          <div
            class="relative z-10 p-8 bg-white bg-opacity-95 rounded-3xl border shadow-2xl backdrop-blur-sm border-white/20 dark:bg-slate-900 dark:border-slate-700/50"
            style="box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25), 0 0 0 1px rgba(255, 255, 255, 0.8)"
          >
                      <div>
              <div class="flex justify-center items-center mb-8">
                <div class="relative">
                  <div class="flex justify-center items-center w-20 h-20 bg-gradient-to-br from-blue-500 to-indigo-600 rounded-2xl shadow-lg transition-transform duration-300 transform rotate-3 hover:rotate-6">
                    <img class="w-12 h-12" :src="$GIN_VUE_ADMIN.appLogo" alt="Logo" />
                  </div>
                  <div class="flex absolute -top-2 -right-2 justify-center items-center w-6 h-6 bg-gradient-to-br from-emerald-400 to-emerald-500 rounded-full shadow-lg">
                    <svg class="w-3 h-3 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path>
                    </svg>
                  </div>
                </div>
              </div>
                              <div class="mb-8 text-center">
                  <h1 class="mb-2 text-3xl font-bold text-gray-800 dark:text-white">
                    {{ $GIN_VUE_ADMIN.appName }}
                  </h1>
                  <p class="text-sm text-gray-500 dark:text-gray-300">
                    NESMA 知识管理系统
                  </p>
                  <div class="flex justify-center items-center mt-4 space-x-2">
                    <div class="w-2 h-2 bg-blue-400 rounded-full opacity-60"></div>
                    <div class="w-12 h-1 bg-gradient-to-r from-blue-500 to-indigo-600 rounded-full"></div>
                    <div class="w-2 h-2 bg-indigo-400 rounded-full opacity-60"></div>
                  </div>
                </div>
            <el-form
              ref="loginForm"
              :model="loginFormData"
              :rules="rules"
              :validate-on-rule-change="false"
              @keyup.enter="submitForm"
              class="space-y-6"
            >
              <el-form-item prop="username">
                <el-input
                  v-model="loginFormData.username"
                  size="large"
                  placeholder="请输入用户名"
                  prefix-icon="User"
                  class="h-12 input-with-shadow"
                />
              </el-form-item>
              <el-form-item prop="password">
                <el-input
                  v-model="loginFormData.password"
                  show-password
                  size="large"
                  type="password"
                  placeholder="请输入密码"
                  prefix-icon="Lock"
                  class="h-12 input-with-shadow"
                />
              </el-form-item>
              <el-form-item
                v-if="loginFormData.openCaptcha"
                prop="captcha"
              >
                <div class="flex gap-3 w-full">
                  <el-input
                    v-model="loginFormData.captcha"
                    placeholder="请输入验证码"
                    size="large"
                    class="flex-1 h-12 input-with-shadow"
                  />
                  <div class="flex justify-center items-center w-28 h-12 bg-gradient-to-r from-blue-50 to-indigo-50 rounded-lg border border-blue-200 transition-all cursor-pointer hover:from-blue-100 hover:to-indigo-100 hover:border-blue-300 hover:shadow-md">
                    <img
                      v-if="picPath"
                      class="object-cover w-full h-full rounded-lg"
                      :src="picPath"
                      alt="验证码"
                      @click="loginVerify()"
                    />
                    <span v-else class="text-xs text-blue-500">点击获取</span>
                  </div>
                </div>
              </el-form-item>
                            <el-form-item class="pt-4">
                <el-button
                  class="w-full h-12 text-base font-semibold bg-gradient-to-r from-blue-500 to-indigo-600 border-none shadow-lg transition-all duration-300 transform hover:from-blue-600 hover:to-indigo-700 hover:shadow-xl hover:scale-105"
                  type="primary"
                  size="large"
                  @click="submitForm"
                  >登 录</el-button
                >
              </el-form-item>
            </el-form>
          </div>
        </div>
      </div>
      
      <!-- 简洁的版权信息 -->
      <div class="absolute right-0 left-0 bottom-6 text-center">
        <div class="flex justify-center items-center space-x-1 text-sm text-gray-500">
          <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
            <path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z"></path>
          </svg>
          <span>© 2025 NESMA 知识管理系统</span>
          <span class="mx-2">|</span>
          <span class="text-blue-500">All rights reserved</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
  import { captcha } from '@/api/user'
  import { checkDB } from '@/api/initdb'
  import { reactive, ref } from 'vue'
  import { ElMessage } from 'element-plus'
  import { useRouter } from 'vue-router'
  import { useUserStore } from '@/pinia/modules/user'

  defineOptions({
    name: 'Login'
  })

  const router = useRouter()
  // 验证函数
  const checkUsername = (rule, value, callback) => {
    if (value.length < 5) {
      return callback(new Error('请输入正确的用户名'))
    } else {
      callback()
    }
  }
  const checkPassword = (rule, value, callback) => {
    if (value.length < 6) {
      return callback(new Error('请输入正确的密码'))
    } else {
      callback()
    }
  }

  // 获取验证码
  const loginVerify = async () => {
    const ele = await captcha()
    rules.captcha.push({
      max: ele.data.captchaLength,
      min: ele.data.captchaLength,
      message: `请输入${ele.data.captchaLength}位验证码`,
      trigger: 'blur'
    })
    picPath.value = ele.data.picPath
    loginFormData.captchaId = ele.data.captchaId
    loginFormData.openCaptcha = ele.data.openCaptcha
  }
  loginVerify()

  // 登录相关操作
  const loginForm = ref(null)
  const picPath = ref('')
  const loginFormData = reactive({
    username: 'admin',
    password: '',
    captcha: '',
    captchaId: '',
    openCaptcha: false
  })
  const rules = reactive({
    username: [{ validator: checkUsername, trigger: 'blur' }],
    password: [{ validator: checkPassword, trigger: 'blur' }],
    captcha: [
      {
        message: '验证码格式不正确',
        trigger: 'blur'
      }
    ]
  })

  const userStore = useUserStore()
  const login = async () => {
    return await userStore.LoginIn(loginFormData)
  }
  const submitForm = () => {
    loginForm.value.validate(async (v) => {
      if (!v) {
        // 未通过前端静态验证
        ElMessage({
          type: 'error',
          message: '请正确填写登录信息',
          showClose: true
        })
        await loginVerify()
        return false
      }

      // 通过验证，请求登陆
      const flag = await login()

      // 登陆失败，刷新验证码
      if (!flag) {
        await loginVerify()
        return false
      }

      // 登陆成功
      return true
    })
  }

  // 跳转初始化
  const checkInit = async () => {
    const res = await checkDB()
    if (res.code === 0) {
      if (res.data?.needInit) {
        userStore.NeedInit()
        await router.push({ name: 'Init' })
      } else {
        ElMessage({
          type: 'info',
          message: '已配置数据库信息，无法初始化'
        })
      }
    }
  }
</script>

<style scoped>
  .input-with-shadow :deep(.el-input__wrapper) {
    box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);
    border: 1px solid #e5e7eb;
    border-radius: 12px;
    transition: all 0.3s ease;
  }
  
  .input-with-shadow :deep(.el-input__wrapper:hover) {
    box-shadow: 0 8px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05);
    border-color: #3b82f6;
  }
  
  .input-with-shadow :deep(.el-input__wrapper.is-focus) {
    box-shadow: 0 8px 15px -3px rgba(59, 130, 246, 0.15), 0 4px 6px -2px rgba(59, 130, 246, 0.1);
    border-color: #3b82f6;
  }
  
  .input-with-shadow :deep(.el-input__inner) {
    color: #374151;
    font-weight: 500;
  }
  
  .input-with-shadow :deep(.el-input__inner::placeholder) {
    color: #9ca3af;
    font-weight: 400;
  }
  
  .input-with-shadow :deep(.el-input__prefix-inner) {
    color: #6b7280;
  }
</style>
