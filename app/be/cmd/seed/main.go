package main

func main() {
	startTerritorySeeder()

	// This is not necessary now because UpdateMarketPrices repo function will create initial listing
	// startMarketSeeder()

	// This is not necessary now because repo reads from yaml directly
	// startOperationsSeeder()
}
