// src/config/navigationConfig.ts

export interface NavigationItem {
  id: string;
  name: string;
  path: string;
  icon: string;
  tooltip: string;
  requiresRegion: boolean;
  priority?: number; // For mobile menu ordering
  showInHeader?: boolean;
  showInSidebar?: boolean;
  showInMobile?: boolean;
  mobileGroup?: 'primary' | 'secondary'; // For grouping mobile menu items
}

export const navigationConfig: NavigationItem[] = [
  {
    id: 'dashboard',
    name: 'Dashboard',
    path: '/',
    icon: '📊',
    tooltip: 'View your empire overview and quick stats',
    requiresRegion: false,
    priority: 1,
    showInHeader: true,
    showInSidebar: true,
    showInMobile: true,
    mobileGroup: 'primary'
  },
  {
    id: 'campaigns',
    name: 'Campaign',
    path: '/campaigns',
    icon: '📜',
    tooltip: 'Start or continue your criminal campaigns',
    requiresRegion: true,
    priority: 2,
    showInHeader: true,
    showInSidebar: true,
    showInMobile: false, // May not be shown in mobile due to space
    mobileGroup: 'secondary'
  },
  {
    id: 'travel',
    name: 'Travel',
    path: '/travel',
    icon: '✈️',
    tooltip: 'Travel to different regions to expand your empire',
    requiresRegion: false,
    priority: 8,
    showInHeader: true,
    showInSidebar: true,
    showInMobile: true,
    mobileGroup: 'primary'
  },
  {
    id: 'territory',
    name: 'Territory',
    path: '/territory',
    icon: '🏙️',
    tooltip: 'Manage and expand your controlled territory',
    requiresRegion: true,
    priority: 3,
    showInHeader: true,
    showInSidebar: true,
    showInMobile: true,
    mobileGroup: 'primary'
  },
  {
    id: 'operations',
    name: 'Operations',
    path: '/operations',
    icon: '🎯',
    tooltip: 'Launch criminal operations for money and resources',
    requiresRegion: true,
    priority: 4,
    showInHeader: true,
    showInSidebar: true,
    showInMobile: true,
    mobileGroup: 'primary'
  },
  {
    id: 'market',
    name: 'Market',
    path: '/market',
    icon: '💹',
    tooltip: 'Buy and sell crew, weapons, and vehicles',
    requiresRegion: true,
    priority: 5,
    showInHeader: true,
    showInSidebar: true,
    showInMobile: true,
    mobileGroup: 'primary'
  },
  {
    id: 'rankings',
    name: 'Rankings',
    path: '/rankings',
    icon: '🏆',
    tooltip: 'See how you rank against other crime lords',
    requiresRegion: false,
    priority: 6,
    showInHeader: true,
    showInSidebar: true,
    showInMobile: false, // May not be shown in mobile due to space
    mobileGroup: 'secondary'
  },
  {
    id: 'nft',
    name: 'NFT',
    path: '/nft',
    icon: '💎',
    tooltip: 'Collect and trade unique digital assets',
    requiresRegion: false,
    priority: 7,
    showInHeader: true,
    showInSidebar: true,
    showInMobile: false, // May not be shown in mobile due to space
    mobileGroup: 'secondary'
  }
];

// Utility functions to get filtered navigation items
export const getHeaderNavItems = () => {
  return navigationConfig
    .filter(item => item.showInHeader)
    .sort((a, b) => (a.priority || 0) - (b.priority || 0));
};

export const getSidebarNavItems = () => {
  return navigationConfig
    .filter(item => item.showInSidebar)
    .sort((a, b) => (a.priority || 0) - (b.priority || 0));
};

export const getMobileNavItems = () => {
  // For mobile, we'll show primary items and manage overflow with horizontal scroll
  return navigationConfig
    .filter(item => item.showInMobile)
    .sort((a, b) => (a.priority || 0) - (b.priority || 0));
};

export const getSecondaryMobileNavItems = () => {
  // Secondary items for an overflow menu or "More" section
  return navigationConfig
    .filter(item => item.mobileGroup === 'secondary' && item.showInMobile === false)
    .sort((a, b) => (a.priority || 0) - (b.priority || 0));
};
