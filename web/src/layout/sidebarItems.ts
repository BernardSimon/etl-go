import {
  DatabaseOutlined,
  SettingOutlined,
  ScheduleOutlined,
  ClockCircleOutlined,
} from "@ant-design/icons-vue";
import type { SidebarItem } from "../types";

// 侧边栏菜单
export const sidebarItems: SidebarItem[] = [
  {
    index: "/datasource",
    title: "router.datasource",
    icon: DatabaseOutlined,
  },

  {
    index: "/workflow-management",
    title: "router.task",
    icon: ScheduleOutlined,
  },
  {
    index: "/system-variables",
    title: "router.systemVariable",
    icon: SettingOutlined,
  },

  {
    index: "/run-logs",
    title: "router.runLog",
    icon: ClockCircleOutlined,
  },
  {
    index: "/files",
    title: "router.file",
    icon: ClockCircleOutlined,
  },
];
