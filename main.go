package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"code.cloudfoundry.org/lager"
	"github.com/monostream/helmi/pkg/broker"
	"github.com/monostream/helmi/pkg/catalog"
	"github.com/monostream/helmi/pkg/config"
	"github.com/monostream/helmi/pkg/helm"
)

func main() {
	configuration := &config.Config{}
	configuration.LoadConfig()

	logger := lager.NewLogger("helmi")
	logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.DEBUG))
	logger.RegisterSink(lager.NewWriterSink(os.Stderr, lager.ERROR))

	c, err := catalog.New(configuration.CatalogPath)

	if err != nil {
		log.Fatal("Failed to parse catalog. Did you set CATALOG_URL correctly? Error:", err)
	}

	err = verifyChartVersions(c)

	if err != nil {
		log.Fatal(err)
	}

	if configuration.Username == "" || configuration.Password == "" {
		log.Println("Username and/or password not specified, authentication will be disabled!")
	}

	b := broker.NewBroker(c, configuration, logger)
	b.Run()
}

func verifyChartVersions(catalog *catalog.Catalog) error {
	charts, err := helm.ListCharts()

	if err != nil {
		return err
	}

	for _, service := range catalog.Services() {
		for _, plan := range service.Plans {
			chartName := service.Chart
			chartVersion := service.ChartVersion

			if len(plan.Chart) > 0 {
				chartName = plan.Chart
			}

			if len(plan.ChartVersion) > 0 {
				chartVersion = plan.ChartVersion
			}

			if chart, ok := charts[chartName]; ok {
				if !strings.EqualFold(chart.ChartVersion, chartVersion) {
					fmt.Println(fmt.Sprintf("Outdated Chart %v: %v <> %v", chartName, chartVersion, chart.ChartVersion))
				}
			} else {
				fmt.Println(fmt.Sprintf("Missing Chart %v", chartName))
			}
		}
	}

	return nil
}
