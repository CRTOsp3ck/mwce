// be/internal/service/campaign_loader.go

package service

import (
	"fmt"
	"mwce-be/internal/model"
	"mwce-be/internal/repository"
	"os"
	"time"

	"github.com/rs/zerolog"
	"gopkg.in/yaml.v3"
)

// CampaignData represents the structure of the campaigns.yaml file
type CampaignData struct {
	Campaigns []CampaignTemplate `yaml:"campaigns"`
}

// CampaignTemplate represents a campaign template in the YAML file
type CampaignTemplate struct {
	ID          string            `yaml:"id"`
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	ImageURL    string            `yaml:"image_url,omitempty"`
	IsActive    bool              `yaml:"is_active"`
	Chapters    []ChapterTemplate `yaml:"chapters"`
}

// ChapterTemplate represents a chapter template in the YAML file
type ChapterTemplate struct {
	ID          string            `yaml:"id"`
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Order       int               `yaml:"order"`
	Missions    []MissionTemplate `yaml:"missions"`
}

// MissionTemplate represents a mission template in the YAML file
type MissionTemplate struct {
	ID            string           `yaml:"id"`
	Name          string           `yaml:"name"`
	Description   string           `yaml:"description"`
	Order         int              `yaml:"order"`
	Prerequisites []string         `yaml:"prerequisites,omitempty"`
	Branches      []BranchTemplate `yaml:"branches"`
}

// BranchTemplate represents a branch template in the YAML file
type BranchTemplate struct {
	ID          string              `yaml:"id"`
	Name        string              `yaml:"name"`
	Description string              `yaml:"description"`
	Operations  []OperationTemplate `yaml:"operations,omitempty"`
	POIs        []POITemplate       `yaml:"pois,omitempty"`
}

// **NOTE: Already present in operations loader**
// OperationTemplate represents a campaign operation template in the YAML file
// type OperationTemplate struct {
// 	ID           string                    `yaml:"id"`
// 	Name         string                    `yaml:"name"`
// 	Description  string                    `yaml:"description"`
// 	Type         string                    `yaml:"type"`
// 	IsSpecial    bool                      `yaml:"is_special"`
// 	Regions      []string                  `yaml:"regions,omitempty"`
// 	Requirements OperationRequirementsYAML `yaml:"requirements"`
// 	Resources    OperationResourcesYAML    `yaml:"resources"`
// 	Rewards      OperationRewardsYAML      `yaml:"rewards"`
// 	Risks        OperationRisksYAML        `yaml:"risks"`
// 	Duration     int                       `yaml:"duration"`
// 	SuccessRate  int                       `yaml:"success_rate"`
// }

// POITemplate represents a campaign POI template in the YAML file
type POITemplate struct {
	ID           string             `yaml:"id"`
	Name         string             `yaml:"name"`
	Description  string             `yaml:"description"`
	Type         string             `yaml:"type"`
	BusinessType string             `yaml:"business_type"`
	IsLegal      bool               `yaml:"is_legal"`
	CityID       string             `yaml:"city_id"`
	Dialogues    []DialogueTemplate `yaml:"dialogues,omitempty"`
}

// DialogueTemplate represents a dialogue template in the YAML file
type DialogueTemplate struct {
	ID              string                 `yaml:"id"`
	Speaker         string                 `yaml:"speaker"`
	InteractionType string                 `yaml:"interaction_type,omitempty"`
	Text            string                 `yaml:"text"`
	Order           int                    `yaml:"order"`
	IsSuccess       *bool                  `yaml:"is_success,omitempty"`
	ResourceEffect  ResourceEffectTemplate `yaml:"resource_effect,omitempty"`
}

// ResourceEffectTemplate represents a resource effect template in the YAML file
type ResourceEffectTemplate struct {
	Money     int `yaml:"money,omitempty"`
	Crew      int `yaml:"crew,omitempty"`
	Weapons   int `yaml:"weapons,omitempty"`
	Vehicles  int `yaml:"vehicles,omitempty"`
	Respect   int `yaml:"respect,omitempty"`
	Influence int `yaml:"influence,omitempty"`
	Heat      int `yaml:"heat,omitempty"`
}

// LoadCampaignData loads campaign data from YAML and seeds it into the database
func LoadCampaignData(campaignRepo repository.CampaignRepository, logger zerolog.Logger) error {
	// Get the campaigns YAML file path
	campaignsFile := os.Getenv("CAMPAIGNS_FILE")
	if campaignsFile == "" {
		campaignsFile = "./configs/campaigns.yaml"
	}

	// Check if file exists
	if _, err := os.Stat(campaignsFile); os.IsNotExist(err) {
		logger.Warn().Str("file", campaignsFile).Msg("Campaigns file not found, skipping seed")
		return nil
	}

	// Read the file
	data, err := os.ReadFile(campaignsFile)
	if err != nil {
		return fmt.Errorf("failed to read campaigns file: %w", err)
	}

	// Parse the YAML
	var campaignData CampaignData
	if err := yaml.Unmarshal(data, &campaignData); err != nil {
		return fmt.Errorf("failed to parse campaigns data: %w", err)
	}

	// Check if campaigns already exist
	var count int64
	if err := campaignRepo.GetDB().Model(&model.Campaign{}).Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		logger.Info().Int64("count", count).Msg("Campaigns already exist, skipping seed")
		return nil
	}

	// Begin transaction
	tx := campaignRepo.GetDB().Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// Rollback on error
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	now := time.Now()

	// Seed campaigns
	for _, campaignTemplate := range campaignData.Campaigns {
		campaign := model.Campaign{
			ID:          campaignTemplate.ID,
			Name:        campaignTemplate.Name,
			Description: campaignTemplate.Description,
			ImageURL:    campaignTemplate.ImageURL,
			IsActive:    campaignTemplate.IsActive,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		if err := tx.Create(&campaign).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create campaign: %w", err)
		}

		// Seed chapters
		for _, chapterTemplate := range campaignTemplate.Chapters {
			chapter := model.Chapter{
				ID:          chapterTemplate.ID,
				CampaignID:  campaign.ID,
				Name:        chapterTemplate.Name,
				Description: chapterTemplate.Description,
				Order:       chapterTemplate.Order,
				CreatedAt:   now,
				UpdatedAt:   now,
			}

			if err := tx.Create(&chapter).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create chapter: %w", err)
			}

			// Seed missions
			for _, missionTemplate := range chapterTemplate.Missions {
				mission := model.Mission{
					ID:            missionTemplate.ID,
					ChapterID:     chapter.ID,
					Name:          missionTemplate.Name,
					Description:   missionTemplate.Description,
					Order:         missionTemplate.Order,
					Prerequisites: missionTemplate.Prerequisites,
					CreatedAt:     now,
					UpdatedAt:     now,
				}

				if err := tx.Create(&mission).Error; err != nil {
					tx.Rollback()
					return fmt.Errorf("failed to create mission: %w", err)
				}

				// Seed branches
				for _, branchTemplate := range missionTemplate.Branches {
					branch := model.Branch{
						ID:          branchTemplate.ID,
						MissionID:   mission.ID,
						Name:        branchTemplate.Name,
						Description: branchTemplate.Description,
						CreatedAt:   now,
						UpdatedAt:   now,
					}

					if err := tx.Create(&branch).Error; err != nil {
						tx.Rollback()
						return fmt.Errorf("failed to create branch: %w", err)
					}

					// Seed operations
					for _, operationTemplate := range branchTemplate.Operations {
						// Convert regions to region IDs
						var regionIDs []string
						for _, regionName := range operationTemplate.Regions {
							var region model.Region
							if err := tx.Where("name = ?", regionName).First(&region).Error; err != nil {
								logger.Warn().Str("region", regionName).Msg("Region not found, skipping")
								continue
							}
							regionIDs = append(regionIDs, region.ID)
						}

						operation := model.CampaignOperation{
							ID:          operationTemplate.ID,
							BranchID:    branch.ID,
							Name:        operationTemplate.Name,
							Description: operationTemplate.Description,
							Type:        operationTemplate.Type,
							IsSpecial:   operationTemplate.IsSpecial,
							RegionIDs:   regionIDs,
							Requirements: model.OperationRequirements{
								MinInfluence:         operationTemplate.Requirements.MinInfluence,
								MaxHeat:              operationTemplate.Requirements.MaxHeat,
								MinTitle:             operationTemplate.Requirements.MinTitle,
								RequiredHotspotTypes: operationTemplate.Requirements.RequiredHotspotTypes,
							},
							Resources: model.OperationResources{
								Crew:     operationTemplate.Resources.Crew,
								Weapons:  operationTemplate.Resources.Weapons,
								Vehicles: operationTemplate.Resources.Vehicles,
								Money:    operationTemplate.Resources.Money,
							},
							Rewards: model.OperationRewards{
								Money:         operationTemplate.Rewards.Money,
								Crew:          operationTemplate.Rewards.Crew,
								Weapons:       operationTemplate.Rewards.Weapons,
								Vehicles:      operationTemplate.Rewards.Vehicles,
								Respect:       operationTemplate.Rewards.Respect,
								Influence:     operationTemplate.Rewards.Influence,
								HeatReduction: operationTemplate.Rewards.HeatReduction,
							},
							Risks: model.OperationRisks{
								CrewLoss:     operationTemplate.Risks.CrewLoss,
								WeaponsLoss:  operationTemplate.Risks.WeaponsLoss,
								VehiclesLoss: operationTemplate.Risks.VehiclesLoss,
								MoneyLoss:    operationTemplate.Risks.MoneyLoss,
								HeatIncrease: operationTemplate.Risks.HeatIncrease,
								RespectLoss:  operationTemplate.Risks.RespectLoss,
							},
							Duration:    operationTemplate.Duration,
							SuccessRate: operationTemplate.SuccessRate,
							CreatedAt:   now,
							UpdatedAt:   now,
						}

						if err := tx.Create(&operation).Error; err != nil {
							tx.Rollback()
							return fmt.Errorf("failed to create operation: %w", err)
						}
					}

					// Seed POIs
					for _, poiTemplate := range branchTemplate.POIs {
						poi := model.CampaignPOI{
							ID:           poiTemplate.ID,
							BranchID:     branch.ID,
							Name:         poiTemplate.Name,
							Description:  poiTemplate.Description,
							Type:         poiTemplate.Type,
							BusinessType: poiTemplate.BusinessType,
							IsLegal:      poiTemplate.IsLegal,
							CityID:       poiTemplate.CityID,
							CreatedAt:    now,
							UpdatedAt:    now,
						}

						if err := tx.Create(&poi).Error; err != nil {
							tx.Rollback()
							return fmt.Errorf("failed to create POI: %w", err)
						}

						// Seed dialogues
						for _, dialogueTemplate := range poiTemplate.Dialogues {
							// Convert interaction type
							var interactionType *model.InteractionType
							if dialogueTemplate.InteractionType != "" {
								it := model.InteractionType(dialogueTemplate.InteractionType)
								interactionType = &it
							}

							dialogue := model.Dialogue{
								ID:              dialogueTemplate.ID,
								POIID:           poi.ID,
								Speaker:         dialogueTemplate.Speaker,
								InteractionType: interactionType,
								Text:            dialogueTemplate.Text,
								Order:           dialogueTemplate.Order,
								IsSuccess:       dialogueTemplate.IsSuccess,
								ResourceEffect: model.ResourceEffect{
									Money:     dialogueTemplate.ResourceEffect.Money,
									Crew:      dialogueTemplate.ResourceEffect.Crew,
									Weapons:   dialogueTemplate.ResourceEffect.Weapons,
									Vehicles:  dialogueTemplate.ResourceEffect.Vehicles,
									Respect:   dialogueTemplate.ResourceEffect.Respect,
									Influence: dialogueTemplate.ResourceEffect.Influence,
									Heat:      dialogueTemplate.ResourceEffect.Heat,
								},
								CreatedAt: now,
								UpdatedAt: now,
							}

							if err := tx.Create(&dialogue).Error; err != nil {
								tx.Rollback()
								return fmt.Errorf("failed to create dialogue: %w", err)
							}
						}
					}
				}
			}
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	logger.Info().Int("campaigns", len(campaignData.Campaigns)).Msg("Campaign data seeded successfully")

	return nil
}

// RunCampaignSeeder seeds campaign data from YAML file
func RunCampaignSeeder(campaignRepo repository.CampaignRepository, logger zerolog.Logger) {
	if err := LoadCampaignData(campaignRepo, logger); err != nil {
		logger.Error().Err(err).Msg("Failed to seed campaign data")
	}
}
