// src/views/NftView.vue

<template>
    <div class="nft-view">
      <div class="page-title">
        <h2>NFT Marketplace</h2>
        <p class="subtitle">Convert your in-game assets to NFTs for trading on the blockchain.</p>
      </div>
      
      <div class="info-banner">
        <div class="banner-icon">ℹ️</div>
        <div class="banner-content">
          <h3>NFT Feature Coming Soon</h3>
          <p>This feature is currently under development. Stay tuned for updates!</p>
        </div>
      </div>
      
      <div class="nft-categories">
        <BaseCard class="category-card">
          <div class="category-icon">🏢</div>
          <h3>Properties</h3>
          <p>Convert your controlled properties into tradable NFTs. Each property NFT generates passive income for its owner.</p>
          <BaseButton variant="secondary" disabled>Coming Soon</BaseButton>
        </BaseCard>
        
        <BaseCard class="category-card">
          <div class="category-icon">👥</div>
          <h3>Crew Members</h3>
          <p>Tokenize your most valuable crew members as NFTs. These character NFTs come with unique skills and bonuses.</p>
          <BaseButton variant="secondary" disabled>Coming Soon</BaseButton>
        </BaseCard>
        
        <BaseCard class="category-card">
          <div class="category-icon">🚗</div>
          <h3>Vehicles</h3>
          <p>Turn your rare and customized vehicles into tradable NFTs. Vehicle NFTs provide special bonuses during operations.</p>
          <BaseButton variant="secondary" disabled>Coming Soon</BaseButton>
        </BaseCard>
        
        <BaseCard class="category-card">
          <div class="category-icon">🔫</div>
          <h3>Weapons</h3>
          <p>Convert your unique weapons into blockchain assets. Weapon NFTs give owners tactical advantages in territory actions.</p>
          <BaseButton variant="secondary" disabled>Coming Soon</BaseButton>
        </BaseCard>
      </div>
      
      <BaseCard class="marketplace-preview">
        <template #header>
          <h3>Marketplace Preview</h3>
        </template>
        
        <div class="marketplace-mockup">
          <div class="mockup-filters">
            <div class="filter-group">
              <label>Category:</label>
              <select disabled>
                <option>All Categories</option>
                <option>Properties</option>
                <option>Crew Members</option>
                <option>Vehicles</option>
                <option>Weapons</option>
              </select>
            </div>
            
            <div class="filter-group">
              <label>Sort By:</label>
              <select disabled>
                <option>Price: Low to High</option>
                <option>Price: High to Low</option>
                <option>Recently Listed</option>
                <option>Popularity</option>
              </select>
            </div>
            
            <div class="filter-group">
              <input type="text" placeholder="Search NFTs..." disabled />
            </div>
          </div>
          
          <div class="nft-grid">
            <div class="nft-item placeholder">
              <div class="item-image"></div>
              <div class="item-details">
                <div class="item-title"></div>
                <div class="item-price"></div>
              </div>
            </div>
            
            <div class="nft-item placeholder">
              <div class="item-image"></div>
              <div class="item-details">
                <div class="item-title"></div>
                <div class="item-price"></div>
              </div>
            </div>
            
            <div class="nft-item placeholder">
              <div class="item-image"></div>
              <div class="item-details">
                <div class="item-title"></div>
                <div class="item-price"></div>
              </div>
            </div>
            
            <div class="nft-item placeholder">
              <div class="item-image"></div>
              <div class="item-details">
                <div class="item-title"></div>
                <div class="item-price"></div>
              </div>
            </div>
            
            <div class="nft-item placeholder">
              <div class="item-image"></div>
              <div class="item-details">
                <div class="item-title"></div>
                <div class="item-price"></div>
              </div>
            </div>
            
            <div class="nft-item placeholder">
              <div class="item-image"></div>
              <div class="item-details">
                <div class="item-title"></div>
                <div class="item-price"></div>
              </div>
            </div>
          </div>
        </div>
        
        <div class="coming-soon-overlay">
          <div class="overlay-content">
            <h3>Marketplace Coming Soon</h3>
            <p>Our NFT marketplace is under development.</p>
            <p>Be the first to know when it launches!</p>
          </div>
        </div>
      </BaseCard>
      
      <BaseCard class="waitlist-card">
        <h3>Join the NFT Waitlist</h3>
        <p>Sign up to be notified when the NFT marketplace is available.</p>
        
        <div class="waitlist-form">
          <input type="email" placeholder="Enter your email" />
          <BaseButton>Join Waitlist</BaseButton>
        </div>
      </BaseCard>
    </div>
  </template>
  
  <script setup lang="ts">
  import BaseCard from '@/components/ui/BaseCard.vue';
  import BaseButton from '@/components/ui/BaseButton.vue';
  </script>
  
  <style lang="scss">
  .nft-view {
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
    
    .info-banner {
      display: flex;
      gap: $spacing-md;
      padding: $spacing-md;
      background-color: rgba($info-color, 0.1);
      border-left: 4px solid $info-color;
      border-radius: $border-radius-md;
      margin-bottom: $spacing-xl;
      
      .banner-icon {
        font-size: 32px;
      }
      
      .banner-content {
        h3 {
          margin: 0 0 $spacing-xs 0;
          color: $info-color;
        }
        
        p {
          margin: 0;
        }
      }
    }
    
    .nft-categories {
      display: grid;
      grid-template-columns: repeat(1, 1fr);
      gap: $spacing-md;
      margin-bottom: $spacing-xl;
      
      @include respond-to(md) {
        grid-template-columns: repeat(2, 1fr);
      }
      
      @include respond-to(lg) {
        grid-template-columns: repeat(4, 1fr);
      }
      
      .category-card {
        @include flex-column;
        align-items: center;
        text-align: center;
        
        .category-icon {
          font-size: 48px;
          margin-bottom: $spacing-md;
        }
        
        h3 {
          margin: 0 0 $spacing-sm 0;
          @include gold-accent;
        }
        
        p {
          margin: 0 0 $spacing-lg 0;
          color: $text-secondary;
        }
      }
    }
    
    .marketplace-preview {
      position: relative;
      margin-bottom: $spacing-xl;
      overflow: hidden;
      
      .marketplace-mockup {
        .mockup-filters {
          display: flex;
          flex-wrap: wrap;
          gap: $spacing-md;
          margin-bottom: $spacing-lg;
          
          .filter-group {
            display: flex;
            align-items: center;
            gap: $spacing-sm;
            
            label {
              font-weight: 500;
              white-space: nowrap;
            }
            
            select, input {
              background-color: $background-lighter;
              color: $text-color;
              border: 1px solid $border-color;
              border-radius: $border-radius-sm;
              padding: 6px 12px;
              opacity: 0.5;
            }
            
            input {
              width: 250px;
            }
          }
        }
        
        .nft-grid {
          display: grid;
          grid-template-columns: repeat(2, 1fr);
          gap: $spacing-md;
          
          @include respond-to(md) {
            grid-template-columns: repeat(3, 1fr);
          }
          
          @include respond-to(lg) {
            grid-template-columns: repeat(6, 1fr);
          }
          
          .nft-item {
            @include card;
            padding: 0;
            overflow: hidden;
            
            &.placeholder {
              .item-image {
                height: 150px;
                background-color: rgba(255, 255, 255, 0.05);
              }
              
              .item-details {
                padding: $spacing-md;
                
                .item-title {
                  height: 16px;
                  width: 80%;
                  background-color: rgba(255, 255, 255, 0.05);
                  margin-bottom: $spacing-sm;
                  border-radius: $border-radius-sm;
                }
                
                .item-price {
                  height: 24px;
                  width: 60%;
                  background-color: rgba(255, 255, 255, 0.05);
                  border-radius: $border-radius-sm;
                }
              }
            }
          }
        }
      }
      
      .coming-soon-overlay {
        position: absolute;
        top: 0;
        left: 0;
        right: 0;
        bottom: 0;
        background-color: rgba($background-color, 0.85);
        @include glass-effect;
        @include flex-center;
        
        .overlay-content {
          text-align: center;
          max-width: 400px;
          padding: $spacing-xl;
          
          h3 {
            @include gold-accent;
            font-size: $font-size-xl;
            margin-bottom: $spacing-md;
          }
          
          p {
            margin: 0 0 $spacing-sm 0;
            font-size: $font-size-lg;
            
            &:last-child {
              margin-bottom: 0;
            }
          }
        }
      }
    }
    
    .waitlist-card {
      text-align: center;
      max-width: 600px;
      margin: 0 auto;
      
      h3 {
        @include gold-accent;
        margin-bottom: $spacing-md;
      }
      
      p {
        margin-bottom: $spacing-lg;
      }
      
      .waitlist-form {
        display: flex;
        gap: $spacing-md;
        
        @include respond-to(sm) {
          flex-direction: row;
        }
        
        @include respond-to(xs) {
          flex-direction: column;
        }
        
        input {
          flex: 1;
          background-color: $background-lighter;
          color: $text-color;
          border: 1px solid $border-color;
          border-radius: $border-radius-sm;
          padding: 8px 16px;
          
          &::placeholder {
            color: $text-secondary;
          }
        }
      }
    }
  }
  </style>