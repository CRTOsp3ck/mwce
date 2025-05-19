# Campaign Mode for Mafia Wars: Criminal Empire

## Core Campaign Concept

The campaign mode would provide narrative-driven gameplay through a series of interconnected missions that utilize your existing operations, territory, and resource systems while adding story choices that create branching paths and multiple endings.

### Key Design Principles

1. **Narrative Integration**: Story content that contextualizes existing game mechanics
2. **Mechanical Consistency**: Campaign missions use the same underlying systems as regular gameplay
3. **Meaningful Choices**: Player decisions that affect both story progression and gameplay rewards
4. **Replayability**: Multiple paths, endings, and difficulty levels to encourage replaying

### Campaign Structure

1. **Campaign**: The overall story container
2. **Chapter**: Major story segments
3. **Mission**: Individual playable story segments allowing for player to choose different branches
4. **Branch**: Available outcome branches for each missions
   - Connected to existing game mechanics (operations and hotspots (as POIs)) for branch completion.
   - Each branch has their own checklist or which operations to do/POIs to interact with.

## Integration with Existing Game Mechanics

### Operations Integration

**Campaign-Specific Operations**: Create operations that only appear during specific campaign missions. Campaign-Specific Operations should be location specific (city-specific) as to make players travel around.

### Hotspot Integration as POI

**Hotspots-as-POIs**: POIs are custom hotspots (think NPCs) that players can interact with to:
1. Have a Dialogue - Dialogues would have various interaction type that would trigger different NPCs responses. Interaction types are like:
- Neutral
- Convince
- Intimidate
These different interaction types would trigger success or failure responses and players lose/gain resources accordingly while the dialogue is ongoing.
2. Extort - Essentially acting as a PvE version of the already in place illegal business but for campaign-specific purposes.
3. Takeover - Essentially acting as a PvE version of the already in place legal business but for campaign-specific purposes.
POIs should also be location specific (city-specific) as to make players travel around just like regular hotspots.

## Story YAML File

Story data should be created in a YAML file in the '\be\configs' folder and seeded into DB when server starts similar to the territory loader (if no seed data is present in DB yet).

## Technical Consideration

Essentially, in the backend, the TerritoryService and OperationsService should have no knowledge of the CampaignService but should just take in CustomProvider interfaces in which the CampaignService should implement and satisty these interface requirements to reduce code coupling. And of course, in the front-end, the TerritoryView.vue and OperationsView.vue should support POIs/operations that are campaign-specific.
