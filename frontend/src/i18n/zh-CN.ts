export default {
  language: { label: "语言", zhCN: "简体中文", zhHK: "繁體中文", enUS: "English" },
  nav: { ask: "提问", explore: "探索", admin: "管理", search: "搜索", toggleTheme: "切换主题" },
  ask: { eyebrow: "匿名提问", defaultTitle: "有什么想问的吗？", description: "请在此写下你的问题。", nickname: "昵称（可选）", content: "想问什么？", attachment: "图片附件（可选）", imageTooLarge: "图片不能超过 {size} KB", submit: "发送问题", submitted: "问题已经投递到收件箱，感谢你的提问。", failed: "提交失败" },
  explore: { eyebrow: "公开探索", title: "探索回答", description: "这里汇集已经回答并公开的问题。", empty: "还没有公开回答。" },
  search: { eyebrow: "公开搜索", title: "搜索问题", placeholder: "输入关键词搜索…", empty: "没有找到相关问题。" },
  question: { anonymous: "匿名", askedAt: "{name} 在 {time} 的提问", answeredAt: "回答于 {time}", unanswered: "暂未回答", attachmentAlt: "提问附图" },
  admin: { eyebrow: "管理后台", title: "管理你的提问箱", loginDescription: "登录后查看收到的问题，并管理公开展示内容。", username: "账号", password: "密码", login: "登录", loginFailed: "登录失败", menuTitle: "你想管理什么？", menuDescription: "选择一个功能开始。", questions: "问题列表", questionsDescription: "查看、回答、发布或删除收到的问题", settings: "自定义设置", settingsDescription: "个性化网站内容与显示选项", dashboard: "仪表盘", allQuestions: "全部问题", pendingQuestions: "待回答" },
  inbox: { eyebrow: "收件箱", title: "问题列表", description: "集中处理收到的问题，回答后可以选择将内容发布到公开展示页。", pending: "待回答", answered: "已回答", published: "已展示", all: "全部问题", answer: "回答", publish: "发布到首页", save: "保存回答", saved: "回答已保存。", empty: "这里暂时没有问题。", loadFailed: "无法读取问题", saveFailed: "保存失败" },
  settings: { eyebrow: "站点外观", title: "自定义设置", description: "逐项调整公开页面的外观和署名。", basics: "基础内容", theme: "主题", siteName: "站点名称", copyrightName: "版权名称", maxUploadKB: "最大图片上传（KB）", primaryColor: "全局主题色", cardOpacity: "卡片透明度：{value}%", save: "保存设置", saved: "设置已保存。", loadFailed: "无法读取设置", saveFailed: "保存失败" },
  error: { requestFailed: "请求失败" },
} as const;
