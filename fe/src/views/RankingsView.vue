// src/views/RankingsView.vue

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import BaseCard from '@/components/ui/BaseCard.vue';
import { usePlayerStore } from '@/stores/modules/player';

const playerStore = usePlayerStore();

// Mock data for rankings
interface PlayerRanking {
  id: string;
  name: string;
  avatar?: string;
  title: string;
  money: number;
  respect: number;
  influence: number;
  territories: number;
  totalTerritories: number;
  rank?: number;
}

// View state
const activeTab = ref<'wealth' | 'respect' | 'influence' | 'territory'>('wealth');
const isLoading = ref(false);
const currentPage = ref(1);
const itemsPerPage = 10;

// Mock rankings data
const wealthRankings = ref<PlayerRanking[]>([]);
const respectRankings = ref<PlayerRanking[]>([]);
const influenceRankings = ref<PlayerRanking[]>([]);
const territoryRankings = ref<PlayerRanking[]>([]);

// Computed properties
const playerMoney = computed(() => playerStore.playerMoney);
const playerRespect = computed(() => playerStore.playerRespect);
const playerInfluence = computed(() => playerStore.playerInfluence);
const controlledHotspots = computed(() => playerStore.controlledHotspots);
const totalHotspots = computed(() => playerStore.totalHotspots);
const playerId = computed(() => playerStore.profile?.id || '');

const currentRankings = computed(() => {
  let rankings: PlayerRanking[] = [];

  switch (activeTab.value) {
    case 'wealth':
      rankings = wealthRankings.value;
      break;
    case 'respect':
      rankings = respectRankings.value;
      break;
    case 'influence':
      rankings = influenceRankings.value;
      break;
    case 'territory':
      rankings = territoryRankings.value;
      break;
  }

  // Apply pagination
  const startIndex = (currentPage.value - 1) * itemsPerPage;
  const endIndex = startIndex + itemsPerPage;

  return rankings.slice(startIndex, endIndex);
});

const totalPages = computed(() => {
  let total = 0;

  switch (activeTab.value) {
    case 'wealth':
      total = Math.ceil(wealthRankings.value.length / itemsPerPage);
      break;
    case 'respect':
      total = Math.ceil(respectRankings.value.length / itemsPerPage);
      break;
    case 'influence':
      total = Math.ceil(influenceRankings.value.length / itemsPerPage);
      break;
    case 'territory':
      total = Math.ceil(territoryRankings.value.length / itemsPerPage);
      break;
  }

  return Math.max(1, total);
});

// Load data when component is mounted
onMounted(async () => {
  if (!playerStore.profile) {
    await playerStore.fetchProfile();
  }

  await refreshRankings();
});

// Helper functions
function formatNumber(value: number): string {
  if (value >= 1000000) {
    return (value / 1000000).toFixed(1) + 'M';
  } else if (value >= 1000) {
    return (value / 1000).toFixed(1) + 'K';
  }
  return value.toString();
}

function formatValue(player: PlayerRanking, type: string): string {
  switch (type) {
    case 'wealth':
      return '$' + formatNumber(player.money);
    case 'respect':
      return player.respect + '%';
    case 'influence':
      return player.influence + '%';
    case 'territory':
      return `${player.territories}/${player.totalTerritories}`;
    default:
      return '';
  }
}

function getInitials(name: string): string {
  return name
    .split(' ')
    .map(n => n[0])
    .join('')
    .toUpperCase()
    .substring(0, 2);
}

function getRankEmoji(rank: number): string {
  switch (rank) {
    case 1:
      return '🥇';
    case 2:
      return '🥈';
    case 3:
      return '🥉';
    default:
      return '';
  }
}

function getRankingTitle(): string {
  switch (activeTab.value) {
    case 'wealth':
      return 'Wealthiest Criminals';
    case 'respect':
      return 'Most Respected Criminals';
    case 'influence':
      return 'Most Influential Criminals';
    case 'territory':
      return 'Largest Criminal Empires';
    default:
      return 'Criminal Rankings';
  }
}

function getValueHeader(): string {
  switch (activeTab.value) {
    case 'wealth':
      return 'Money';
    case 'respect':
      return 'Respect';
    case 'influence':
      return 'Influence';
    case 'territory':
      return 'Territories';
    default:
      return 'Value';
  }
}

function isCurrentPlayer(id: string): boolean {
  return id === playerId.value;
}

function getCurrentPlayerRanking(type?: string): PlayerRanking | undefined {
  const rankingType = type || activeTab.value;
  let rankings: PlayerRanking[] = [];

  switch (rankingType) {
    case 'wealth':
      rankings = wealthRankings.value;
      break;
    case 'respect':
      rankings = respectRankings.value;
      break;
    case 'influence':
      rankings = influenceRankings.value;
      break;
    case 'territory':
      rankings = territoryRankings.value;
      break;
  }

  return rankings.find(p => p.id === playerId.value);
}

function getTopValue(type: string): number | string {
  let rankings: PlayerRanking[] = [];

  switch (type) {
    case 'wealth':
      rankings = wealthRankings.value;
      return rankings.length > 0 ? rankings[0].money : 0;
    case 'respect':
      rankings = respectRankings.value;
      return rankings.length > 0 ? rankings[0].respect : 0;
    case 'influence':
      rankings = influenceRankings.value;
      return rankings.length > 0 ? rankings[0].influence : 0;
    case 'territory':
      rankings = territoryRankings.value;
      return rankings.length > 0 ? `${rankings[0].territories}/${rankings[0].totalTerritories}` : '0/0';
    default:
      return 0;
  }
}

function prevPage(): void {
  if (currentPage.value > 1) {
    currentPage.value--;
  }
}

function nextPage(): void {
  if (currentPage.value < totalPages.value) {
    currentPage.value++;
  }
}

async function refreshRankings(): Promise<void> {
  isLoading.value = true;

  try {
    // Simulate API call to get rankings
    await new Promise(resolve => setTimeout(resolve, 1000));

    // Mock data generation
    generateMockRankings();
  } finally {
    isLoading.value = false;
  }
}

function generateMockRankings(): void {
  // Generate some mock players
  const playerNames = [
    'Don Corleone', 'Tony Montana', 'Lucky Luciano', 'Al Capone', 'Bugsy Siegel',
    'Meyer Lansky', 'Frank Costello', 'Carlo Gambino', 'John Gotti', 'Pablo Escobar',
    'Griselda Blanco', 'Frank Lucas', 'Whitey Bulger', 'Dutch Schultz', 'Bumpy Johnson',
    'Nicky Barnes', 'Mickey Cohen', 'Hyman Roth', 'Moe Greene', 'Sonny Black',
    'Jimmy Conway', 'Tommy DeVito', 'Henry Hill', 'Paulie Cicero', 'Johnny Roselli'
  ];

  const playerTitles = [
    'Associate', 'Soldier', 'Capo', 'Underboss', 'Consigliere', 'Boss', 'Godfather'
  ];

  // Create a current player entry if a profile exists
  let currentPlayer: PlayerRanking | null = null;

  if (playerStore.profile) {
    currentPlayer = {
      id: playerStore.profile.id,
      name: playerStore.profile.name,
      title: playerStore.profile.title.toString(),
      money: playerStore.profile.money,
      respect: playerStore.profile.respect,
      influence: playerStore.profile.influence,
      territories: playerStore.profile.controlledHotspots,
      totalTerritories: playerStore.profile.totalHotspotCount
    };
  }

  // Generate mock players
  const mockPlayers: PlayerRanking[] = [];

  for (let i = 0; i < 24; i++) {
    const name = playerNames[i];
    const title = playerTitles[Math.floor(Math.random() * playerTitles.length)];

    mockPlayers.push({
      id: `player-${i + 1}`,
      name,
      title,
      money: Math.floor(Math.random() * 1000000) + 100000,
      respect: Math.floor(Math.random() * 100),
      influence: Math.floor(Math.random() * 100),
      territories: Math.floor(Math.random() * 20) + 1,
      totalTerritories: 30
    });
  }

  // Add current player if available
  if (currentPlayer) {
    mockPlayers.push(currentPlayer);
  }

  // Sort players by different criteria
  wealthRankings.value = [...mockPlayers].sort((a, b) => b.money - a.money);
  respectRankings.value = [...mockPlayers].sort((a, b) => b.respect - a.respect);
  influenceRankings.value = [...mockPlayers].sort((a, b) => b.influence - a.influence);
  territoryRankings.value = [...mockPlayers].sort((a, b) => b.territories - a.territories);

  // Add rank to each player
  wealthRankings.value.forEach((player, index) => {
    player.rank = index + 1;
  });

  respectRankings.value.forEach((player, index) => {
    player.rank = index + 1;
  });

  influenceRankings.value.forEach((player, index) => {
    player.rank = index + 1;
  });

  territoryRankings.value.forEach((player, index) => {
    player.rank = index + 1;
  });

  // Reset pagination
  currentPage.value = 1;
}
</script>

<template>
  <div class="rankings-view">
    <div class="page-title">
      <h2>Criminal Rankings</h2>
      <p class="subtitle">See how you stack up against other criminal empires in the city.</p>
    </div>

    <div class="rankings-tabs">
      <button class="tab-button" :class="{ active: activeTab === 'wealth' }" @click="activeTab = 'wealth'">
        Wealth
      </button>
      <button class="tab-button" :class="{ active: activeTab === 'respect' }" @click="activeTab = 'respect'">
        Respect
      </button>
      <button class="tab-button" :class="{ active: activeTab === 'influence' }" @click="activeTab = 'influence'">
        Influence
      </button>
      <button class="tab-button" :class="{ active: activeTab === 'territory' }" @click="activeTab = 'territory'">
        Territory
      </button>
    </div>

    <BaseCard class="rankings-card">
      <template #header>
        <div class="card-header-content">
          <h3>{{ getRankingTitle() }}</h3>
          <div class="refresh-control">
            <button class="refresh-btn" @click="refreshRankings" :disabled="isLoading">
              <span class="refresh-icon" :class="{ 'is-loading': isLoading }">↻</span>
              Refresh
            </button>
          </div>
        </div>
      </template>

      <div class="rankings-table">
        <div class="table-header">
          <div class="header-cell rank">Rank</div>
          <div class="header-cell player">Player</div>
          <div class="header-cell value">{{ getValueHeader() }}</div>
          <div class="header-cell title">Title</div>
        </div>

        <div class="table-body">
          <div v-for="(player, index) in currentRankings" :key="player.id" class="table-row"
            :class="{ 'is-current-player': isCurrentPlayer(player.id) }">
            <div class="cell rank">
              <div class="rank-number">{{ index + 1 }}</div>
              <div v-if="index < 3" class="rank-medal" :class="`medal-${index + 1}`">
                {{ getRankEmoji(index + 1) }}
              </div>
            </div>
            <div class="cell player">
              <div class="player-avatar" v-if="player.avatar">
                <img :src="player.avatar" :alt="player.name" />
              </div>
              <div class="player-avatar placeholder" v-else>
                {{ getInitials(player.name) }}
              </div>
              <div class="player-name">{{ player.name }}</div>
            </div>
            <div class="cell value">{{ formatValue(player, activeTab) }}</div>
            <div class="cell title">{{ player.title }}</div>
          </div>
        </div>
      </div>

      <div v-if="currentRankings.length === 0" class="empty-rankings">
        <p>No rankings available at this time.</p>
      </div>

      <template #footer>
        <div class="rankings-footer">
          <div class="pagination" v-if="currentRankings.length > 0">
            <button class="pagination-btn" :disabled="currentPage === 1" @click="prevPage">
              Previous
            </button>
            <div class="pagination-info">
              Page {{ currentPage }} of {{ totalPages }}
            </div>
            <button class="pagination-btn" :disabled="currentPage === totalPages" @click="nextPage">
              Next
            </button>
          </div>

          <div class="player-ranking" v-if="getCurrentPlayerRanking()">
            <div class="ranking-label">Your Ranking:</div>
            <div class="ranking-value">
              #{{ getCurrentPlayerRanking()?.rank }}
              ({{ formatValue(getCurrentPlayerRanking(), activeTab) }})
            </div>
          </div>
        </div>
      </template>
    </BaseCard>

    <div class="stats-cards">
      <BaseCard class="stats-card">
        <div class="stat-header">
          <div class="stat-icon wealth">💰</div>
          <h3>Wealth Ranking</h3>
        </div>
        <div class="stat-value">#{{ getCurrentPlayerRanking('wealth')?.rank || '?' }}</div>
        <div class="stat-details">
          <div class="detail-item">
            <div class="detail-label">Your Money:</div>
            <div class="detail-value">${{ formatNumber(playerMoney) }}</div>
          </div>
          <div class="detail-item">
            <div class="detail-label">Top Player:</div>
            <div class="detail-value">${{ formatNumber(getTopValue('wealth')) }}</div>
          </div>
        </div>
      </BaseCard>

      <BaseCard class="stats-card">
        <div class="stat-header">
          <div class="stat-icon respect">👊</div>
          <h3>Respect Ranking</h3>
        </div>
        <div class="stat-value">#{{ getCurrentPlayerRanking('respect')?.rank || '?' }}</div>
        <div class="stat-details">
          <div class="detail-item">
            <div class="detail-label">Your Respect:</div>
            <div class="detail-value">{{ playerRespect }}%</div>
          </div>
          <div class="detail-item">
            <div class="detail-label">Top Player:</div>
            <div class="detail-value">{{ getTopValue('respect') }}%</div>
          </div>
        </div>
      </BaseCard>

      <BaseCard class="stats-card">
        <div class="stat-header">
          <div class="stat-icon influence">🏛️</div>
          <h3>Influence Ranking</h3>
        </div>
        <div class="stat-value">#{{ getCurrentPlayerRanking('influence')?.rank || '?' }}</div>
        <div class="stat-details">
          <div class="detail-item">
            <div class="detail-label">Your Influence:</div>
            <div class="detail-value">{{ playerInfluence }}%</div>
          </div>
          <div class="detail-item">
            <div class="detail-label">Top Player:</div>
            <div class="detail-value">{{ getTopValue('influence') }}%</div>
          </div>
        </div>
      </BaseCard>

      <BaseCard class="stats-card">
        <div class="stat-header">
          <div class="stat-icon territory">🏙️</div>
          <h3>Territory Ranking</h3>
        </div>
        <div class="stat-value">#{{ getCurrentPlayerRanking('territory')?.rank || '?' }}</div>
        <div class="stat-details">
          <div class="detail-item">
            <div class="detail-label">Your Territories:</div>
            <div class="detail-value">{{ controlledHotspots }}/{{ totalHotspots }}</div>
          </div>
          <div class="detail-item">
            <div class="detail-label">Top Player:</div>
            <div class="detail-value">{{ getTopValue('territory') }}</div>
          </div>
        </div>
      </BaseCard>
    </div>
  </div>
</template>

<style lang="scss">
.rankings-view {
  .page-title {
    margin-bottom: $spacing-xl;

    h2 {
      @include gold-accent;
      margin-bottom: $spacing-xs;
    }

    .subtitle {
      color: $text-secondary;
    }
  }

  .rankings-tabs {
    display: flex;
    gap: 1px;
    margin-bottom: $spacing-lg;
    background-color: $border-color;
    border-radius: $border-radius-sm;
    overflow: hidden;

    .tab-button {
      flex: 1;
      background-color: $background-lighter;
      border: none;
      padding: $spacing-md;
      color: $text-secondary;
      font-weight: 500;
      cursor: pointer;
      transition: $transition-base;

      &.active {
        background-color: $primary-color;
        color: $text-color;
      }

      &:hover:not(.active) {
        background-color: lighten($background-lighter, 5%);
      }
    }
  }

  .rankings-card {
    margin-bottom: $spacing-xl;

    .card-header-content {
      @include flex-between;
      width: 100%;

      h3 {
        margin: 0;
      }

      .refresh-control {
        .refresh-btn {
          display: flex;
          align-items: center;
          gap: $spacing-xs;
          background: none;
          border: none;
          color: $text-secondary;
          cursor: pointer;
          transition: $transition-base;

          &:hover:not(:disabled) {
            color: $text-color;
          }

          &:disabled {
            opacity: 0.5;
            cursor: not-allowed;
          }

          .refresh-icon {
            display: inline-block;

            &.is-loading {
              animation: spin 1s linear infinite;
            }
          }
        }
      }
    }

    .rankings-table {
      .table-header {
        display: grid;
        grid-template-columns: 80px 1fr 120px 120px;
        gap: $spacing-sm;
        padding: $spacing-sm;
        background-color: $background-lighter;
        border-radius: $border-radius-sm;
        margin-bottom: $spacing-md;

        .header-cell {
          font-weight: 600;

          &.rank {
            text-align: center;
          }
        }
      }

      .table-body {
        @include flex-column;
        gap: $spacing-xs;

        .table-row {
          display: grid;
          grid-template-columns: 80px 1fr 120px 120px;
          gap: $spacing-sm;
          padding: $spacing-sm;
          border-radius: $border-radius-sm;
          transition: $transition-base;

          &:hover {
            background-color: rgba(255, 255, 255, 0.05);
          }

          &.is-current-player {
            background-color: rgba($primary-color, 0.2);
            border-left: 3px solid $primary-color;

            &:hover {
              background-color: rgba($primary-color, 0.3);
            }
          }

          .cell {
            display: flex;
            align-items: center;

            &.rank {
              justify-content: center;
              position: relative;

              .rank-number {
                font-weight: 600;
                font-size: $font-size-lg;
              }

              .rank-medal {
                position: absolute;
                top: -8px;
                right: -8px;
                font-size: 20px;
              }
            }

            &.player {
              .player-avatar {
                width: 36px;
                height: 36px;
                border-radius: 50%;
                margin-right: $spacing-sm;
                overflow: hidden;

                img {
                  width: 100%;
                  height: 100%;
                  object-fit: cover;
                }

                &.placeholder {
                  background-color: $primary-color;
                  color: $text-color;
                  display: flex;
                  align-items: center;
                  justify-content: center;
                  font-weight: 600;
                }
              }

              .player-name {
                font-weight: 500;
              }
            }

            &.value {
              font-weight: 600;
              @include gold-accent;
            }
          }
        }
      }
    }

    .empty-rankings {
      text-align: center;
      padding: $spacing-xl;
      color: $text-secondary;
    }

    .rankings-footer {
      @include flex-between;

      .pagination {
        display: flex;
        align-items: center;
        gap: $spacing-md;

        .pagination-btn {
          background-color: $background-lighter;
          border: none;
          padding: $spacing-xs $spacing-md;
          border-radius: $border-radius-sm;
          color: $text-color;
          cursor: pointer;
          transition: $transition-base;

          &:hover:not(:disabled) {
            background-color: lighten($background-lighter, 5%);
          }

          &:disabled {
            opacity: 0.5;
            cursor: not-allowed;
          }
        }

        .pagination-info {
          color: $text-secondary;
        }
      }

      .player-ranking {
        display: flex;
        align-items: center;
        gap: $spacing-sm;

        .ranking-label {
          color: $text-secondary;
        }

        .ranking-value {
          font-weight: 600;
          @include gold-accent;
        }
      }
    }
  }

  .stats-cards {
    display: grid;
    grid-template-columns: repeat(1, 1fr);
    gap: $spacing-md;

    @include respond-to(sm) {
      grid-template-columns: repeat(2, 1fr);
    }

    @include respond-to(lg) {
      grid-template-columns: repeat(4, 1fr);
    }

    .stats-card {
      @include flex-column;
      align-items: center;
      text-align: center;

      .stat-header {
        @include flex-column;
        align-items: center;
        margin-bottom: $spacing-md;

        .stat-icon {
          font-size: 32px;
          margin-bottom: $spacing-sm;

          &.wealth {
            color: $secondary-color;
          }

          &.respect {
            color: $success-color;
          }

          &.influence {
            color: $info-color;
          }

          &.territory {
            color: $primary-color;
          }
        }

        h3 {
          margin: 0;
          font-size: $font-size-md;
        }
      }

      .stat-value {
        font-size: $font-size-xxl;
        font-weight: 700;
        margin-bottom: $spacing-md;
        @include gold-accent;
      }

      .stat-details {
        width: 100%;

        .detail-item {
          @include flex-between;
          margin-bottom: $spacing-xs;

          &:last-child {
            margin-bottom: 0;
          }

          .detail-label {
            color: $text-secondary;
          }

          .detail-value {
            font-weight: 600;
          }
        }
      }
    }
  }
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
</style>