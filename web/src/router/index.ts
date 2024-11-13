// router.js
import { createRouter, createWebHashHistory, RouteRecordRaw } from "vue-router";

import Channels from "@Components/Channels/index.vue";
import RTCInvite from "@Components/RTCInvite/index.vue";
import RelayInvite from "@Components/RelayInvite/index.vue";
import Settings from "@Components/Settings/index.vue";


export class ROUTE_PATHS {
  public static readonly NONE = "/"
  public static readonly CHANNELS = "/channels"
  public static readonly RTC_INVITE = "/rtc-invite"
  public static readonly RELAY_INVITE = "/relay-invite"
  public static readonly SETTINGS = "/settings"
}

const routes: RouteRecordRaw[] = [
  {
    path: ROUTE_PATHS.NONE,
    component: { template: "<div></div>" }
  },
  {
    path: ROUTE_PATHS.CHANNELS,
    component: Channels
  },
  {
    path: ROUTE_PATHS.RTC_INVITE,
    component: RTCInvite
  },
  {
    path: ROUTE_PATHS.RELAY_INVITE,
    component: RelayInvite
  },
  {
    path: ROUTE_PATHS.SETTINGS,
    component: Settings
  }
];

const router = createRouter({
  history: createWebHashHistory(),
  routes
})

export default router;
