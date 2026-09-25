import { createRouter, createWebHashHistory } from "vue-router";
import AskView from "./views/AskView.vue";
import ExploreView from "./views/ExploreView.vue";
import AdminView from "./views/AdminView.vue";
import InboxView from "./views/InboxView.vue";
import SettingsView from "./views/SettingsView.vue";
import SearchView from "./views/SearchView.vue";

export default createRouter({
  history: createWebHashHistory(),
  routes: [
    { path: "/", redirect: "/ask" },
    { path: "/ask", component: AskView },
    { path: "/explore", component: ExploreView },
    { path: "/search", component: SearchView },
    { path: "/admin", component: AdminView },
    { path: "/admin/questions", component: InboxView },
    { path: "/admin/settings", component: SettingsView },
  ],
});
