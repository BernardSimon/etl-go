import { createApp } from "vue";
import { createPinia } from "pinia";
import App from "./App.vue";
import router from "./router";
import Antd from "ant-design-vue";
import "ant-design-vue/dist/reset.css";
import "./styles/index.scss";
import i18n from "./i18n.ts";
// 创建应用实例
const app = createApp(App);

// 使用 Pinia 状态管理
app.use(createPinia());

// 使用路由
app.use(router);

// 使用 Ant Design Vue
app.use(Antd);
app.use(i18n)

// 挂载应用
app.mount("#app");
