// internal/utils/utils.go
package utils

import (
	"awesomeProject/internal/data"
	"github.com/bytedance/sonic"
	"io/ioutil"
	"log"
	"sync"
)

// LoadCVData läser och parserar JSON-data från en fil
func LoadCVData(filepath string) (*data.CVData, error) {
	bytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	var cv data.CVData
	err = sonic.Unmarshal(bytes, &cv)
	if err != nil {
		return nil, err
	}

	return &cv, nil
}

// ProcessSectionsConcurrently bearbetar olika sektioner av CV-data parallellt
func ProcessSectionsConcurrently(cv *data.CVData) {
	var wg sync.WaitGroup

	sections := []func(){
		func() {
			defer wg.Done()
			// Bearbeta Färdigheter
			log.Println("Bearbetar färdigheter")
			// Här kan du lägga till ytterligare bearbetning om nödvändigt
		},
		func() {
			defer wg.Done()
			// Bearbeta Språk
			log.Println("Bearbetar språk")
			// Här kan du lägga till ytterligare bearbetning om nödvändigt
		},
		func() {
			defer wg.Done()
			// Bearbeta Arbetslivserfarenhet
			log.Println("Bearbetar arbetslivserfarenhet")
			// Här kan du lägga till ytterligare bearbetning om nödvändigt
		},
		// Lägg till fler sektioner efter behov
	}

	wg.Add(len(sections))
	for _, section := range sections {
		go section()
	}

	wg.Wait()
	log.Println("Alla sektioner har bearbetats")
}
