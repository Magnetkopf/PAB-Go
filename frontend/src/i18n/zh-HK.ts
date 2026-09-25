export default {
  language: { label: "語言", zhCN: "簡體中文", zhHK: "繁體中文", enUS: "English" },
  nav: { ask: "提問", explore: "探索", admin: "管理", search: "搜尋", toggleTheme: "切換主題" },
  ask: { eyebrow: "匿名提問", defaultTitle: "有甚麼想問嗎？", description: "請在此寫下你的問題。", nickname: "暱稱（可選）", content: "想問甚麼？", attachment: "圖片附件（可選）", imageTooLarge: "圖片不可超過 {size} KB", submit: "發送問題", submitted: "問題已送到收件箱，感謝你的提問。", failed: "提交失敗" },
  explore: { eyebrow: "公開探索", title: "探索回答", description: "這裡匯集已回答並公開的問題。", empty: "還沒有公開回答。" },
  search: { eyebrow: "公開搜尋", title: "搜尋問題", placeholder: "輸入關鍵字搜尋…", empty: "找不到相符問題。" },
  question: { anonymous: "匿名", askedAt: "{name} 在 {time} 的提問", answeredAt: "回答於 {time}", unanswered: "尚未回答", attachmentAlt: "提問附圖" },
  admin: { eyebrow: "管理後台", title: "管理你的提問箱", loginDescription: "登入後查看收到的問題，並管理公開展示內容。", username: "帳號", password: "密碼", login: "登入", loginFailed: "登入失敗", menuTitle: "你想管理甚麼？", menuDescription: "選擇一項功能開始。", questions: "問題列表", questionsDescription: "查看、回答、發佈或刪除收到的問題", settings: "自訂設定", settingsDescription: "個人化網站內容與顯示選項", dashboard: "儀表板", allQuestions: "全部問題", pendingQuestions: "待回答" },
  inbox: { eyebrow: "收件箱", title: "問題列表", description: "集中處理收到的問題，回答後可選擇將內容發佈至公開展示頁。", pending: "待回答", answered: "已回答", published: "已展示", all: "全部問題", answer: "回答", publish: "發佈到首頁", save: "儲存回答", saved: "回答已儲存。", empty: "這裡暫時沒有問題。", loadFailed: "無法讀取問題", saveFailed: "儲存失敗" },
  settings: { eyebrow: "網站外觀", title: "自訂設定", description: "逐項調整公開頁面的外觀與署名。", basics: "基本內容", theme: "主題", siteName: "網站名稱", copyrightName: "版權名稱", maxUploadKB: "最大圖片上載（KB）", primaryColor: "全域主題色", cardOpacity: "卡片透明度：{value}%", save: "儲存設定", saved: "設定已儲存。", loadFailed: "無法讀取設定", saveFailed: "儲存失敗" },
  error: { requestFailed: "請求失敗" },
} as const;
