<template>
  <div class="h-full bg-slate-50">
    <div class="flex h-full justify-center items-center">
      <div class="h-max min-w-[16rem] w-1/4 max-w-[24rem] text-center items-center">
        <div class="inline-flex mt-4 mb-8 items-center">
          <img src="@/assets/weave.png" class="h-12 mr-2" />
          <h1 class="font-bold text-4xl font-mono">NAS326</h1>
        </div>

        <div v-if="showLogin">
            <el-form ref="loginFormRef" :model="loginUser" size="large" :rules="rules" show-message>
              <el-form-item prop="name">
                <el-input v-model="loginUser.name" placeholder="admin">
                  <template #prefix>
                    <User />
                  </template>
                </el-input>
              </el-form-item>

              <el-form-item prop="password">
                <el-input v-model="loginUser.password" type="password" placeholder="123456" show-password>
                  <template #prefix>
                    <Lock />
                  </template>
                </el-input>
              </el-form-item>
            </el-form>

            <el-button class="w-full" type="primary" size="large" @click="login(loginFormRef)">LOGIN</el-button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
</style>

<script setup>
import { ElMessage, ElNotification } from "element-plus"
import { User, Lock, Github, Wechat } from '@icon-park/vue-next'
import { ref, reactive } from 'vue'
import request from '@/axios'
import { useRouter } from 'vue-router'
import { authInfo } from '@/config.js'

const router = useRouter();

const loginFormRef = ref();
const registerFormRef = ref();
const redirectUri = window.location.origin + '/oauth'

const showLogin = ref(true);

const loginUser = reactive({
  name: "admin",
  password: "123456",
});
const registerUser = reactive({
  name: "",
  email: "",
  password: "",
});
const rules = reactive({
  name: [
    { required: true, message: 'Please input user name', trigger: 'blur' }
  ],
  password: [
    { required: true, message: 'Please input password', trigger: 'blur' },
    { min: 6, message: 'Length should be great than 6', trigger: 'blur' }
  ],
  email: [
    { required: true, message: 'Please input email', trigger: 'blur' },
    { type: 'email', message: 'Please input correct email address', trigger: ['blur', 'change'] },
  ]
});

const login = async (form) => {
  if (!form) {
    return
  }

  let name = loginUser.name;

  let success = function() {
    ElNotification.success({
          title: 'Login Success',
          message: 'Hi~ ' + name,
          showClose: true,
          duration: 1500,
        })
    router.push('/');
  }

  await form.validate((valid, fields) => {
    if (valid) {
      request.post("/api/v1/auth/token", {
        name: loginUser.name,
        password: loginUser.password,
        setCookie: true,
      }).then((response) => {
        success()
      })
    } else {
      console.log("input invaild", fields)
      ElMessage({
        message: "Input invalid" + fields,
        type: "error",
      });
    }
  });
};

const oauthLogin = (authType) => {
  if (!authInfo[authType]) {
    return
  }

  let uri = "";
  const endpoint = authInfo[authType].endpoint;
  const scope = authInfo[authType].scope;
  const clientId = authInfo[authType].clientId;
  const state = btoa(`${window.location.search}&app=weave&oauth=${authType}`)

  if (authType === "google") {
    uri = `${endpoint}?client_id=${clientId}&redirect_uri=${redirectUri}&scope=${scope}&response_type=code&state=${state}`;
  } else if (authType === "github") {
    uri = `${endpoint}?client_id=${clientId}&redirect_uri=${redirectUri}&scope=${scope}&response_type=code&state=${state}`;
  } else if (authType === "wechat") {
    if (navigator.userAgent.includes("MicroMessenger")) {
      uri = `${authInfo[authType].mpEndpoint}?appid=${authInfo[authType].clientId2}&redirect_uri=${redirectUri}&state=${state}&scope=${authInfo[authType].mpScope}&response_type=code#wechat_redirect`;
    } else {
      uri = `${endpoint}?appid=${clientId}&redirect_uri=${redirectUri}&scope=${scope}&response_type=code&state=${state}#wechat_redirect`;
    }
  } else {
    console.log(`auth type ${authType} not supported`)
    return
  }
  window.location.href = uri;
};

const register = async (form) => {
  if (!form) {
    return
  }

  await form.validate((valid, fields) => {
    if (valid) {
      request.post("/api/v1/auth/user", {
        name: registerUser.name,
        password: registerUser.password,
        email: registerUser.email,
      }).then((response) => {
        ElMessage({
          message: 'Register success',
          type: 'success',
        })
        loginUser.name = registerUser.name;
        loginUser.password = registerUser.password;
        activeTab.value = 'login';
      })
    } else {
      console.log("Input invalid =>", fields)
      ElMessage({
        message: "Input invalid",
        type: "error",
      });
    }
  });
};
</script>
