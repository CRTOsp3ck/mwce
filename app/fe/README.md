```
criminal-empire-frontend/
├── public/
│   ├── favicon.ico
│   └── index.html
├── src/
│   ├── assets/
│   │   ├── fonts/
│   │   ├── images/
│   │   └── styles/
│   │       ├── _variables.scss
│   │       ├── _mixins.scss
│   │       └── main.scss
│   ├── components/
│   │   ├── layout/
│   │   │   ├── AppHeader.vue
│   │   │   ├── AppSidebar.vue
│   │   │   └── AppFooter.vue
│   │   ├── ui/
│   │   │   ├── BaseButton.vue
│   │   │   ├── BaseCard.vue
│   │   │   ├── BaseModal.vue
│   │   │   └── BaseNotification.vue
│   │   ├── territory/
│   │   │   ├── TerritoryMap.vue
│   │   │   ├── HotspotCard.vue
│   │   │   └── ActionPanel.vue
│   │   ├── operations/
│   │   │   ├── OperationCard.vue
│   │   │   └── OperationsList.vue
│   │   ├── market/
│   │   │   ├── MarketItem.vue
│   │   │   └── PriceChart.vue
│   │   └── rankings/
│   │       └── LeaderboardTable.vue
│   ├── router/
│   │   ├── index.ts
│   │   └── routes.ts
│   ├── store/
│   │   ├── index.ts
│   │   ├── modules/
│   │   │   ├── player.ts
│   │   │   ├── territory.ts
│   │   │   ├── operations.ts
│   │   │   └── market.ts
│   ├── services/
│   │   ├── api.ts
│   │   ├── playerService.ts
│   │   ├── territoryService.ts
│   │   ├── operationsService.ts
│   │   └── marketService.ts
│   ├── types/
│   │   ├── player.ts
│   │   ├── territory.ts
│   │   ├── operations.ts
│   │   └── market.ts
│   ├── utils/
│   │   ├── formatters.ts
│   │   └── notifications.ts
│   ├── views/
│   │   ├── HomeView.vue
│   │   ├── TerritoryView.vue
│   │   ├── OperationsView.vue
│   │   ├── MarketView.vue
│   │   ├── RankingsView.vue
│   │   └── NftView.vue
│   ├── App.vue
│   ├── main.ts
│   └── shims-vue.d.ts
├── .eslintrc.js
├── .gitignore
├── package.json
├── tsconfig.json
└── vite.config.ts
```