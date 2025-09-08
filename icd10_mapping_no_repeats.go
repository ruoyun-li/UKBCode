package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

type PatientMap map[int][][]int

func dedupeAndSort(codes []int) []int {
	if len(codes) <= 1 {
		if len(codes) == 1 {
			return []int{codes[0]}
		}
		return codes
	}
	set := make(map[int]struct{}, len(codes))
	for _, c := range codes {
		set[c] = struct{}{}
	}
	out := make([]int, 0, len(set))
	for c := range set {
		out = append(out, c)
	}
	sort.Ints(out)
	return out
}

// Given a patient's visits (chronological), keep only codes that were not in the
// immediately previous visit.
func keepNewPerVisit(visits [][]int) [][]int {
	result := make([][]int, 0, len(visits))

	prevSet := map[int]struct{}{}

	for i, v := range visits {
		v = dedupeAndSort(v) // clean current visit

		if i == 0 {
			// First visit: nothing to compare to; keep all
			result = append(result, append([]int(nil), v...))
		} else {
			// Keep only codes not in previous visit
			newCodes := make([]int, 0, len(v))
			for _, c := range v {
				if _, ok := prevSet[c]; !ok {
					newCodes = append(newCodes, c)
				}
			}
			result = append(result, newCodes)
		}

		prevSet = make(map[int]struct{}, len(v))
		for _, c := range v {
			prevSet[c] = struct{}{}
		}
	}
	return result
}

func loadPatientMap(path string) (PatientMap, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var pm PatientMap
	dec := json.NewDecoder(f)
	if err := dec.Decode(&pm); err != nil {
		return nil, err
	}
	return pm, nil
}

func savePatientMap(path string, pm PatientMap) error {
	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(pm); err != nil {
		f.Close()
		os.Remove(tmp)
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run keep_new_dx.go <in_map.json> <out_map.json>")
		os.Exit(1)
	}
	inPath := os.Args[1]
	outPath := os.Args[2]

	pm, err := loadPatientMap(inPath)
	if err != nil {
		fmt.Println("Error loading input:", err)
		os.Exit(1)
	}

	if pm == nil {
		fmt.Println("Input map is empty or invalid.")
		os.Exit(1)
	}

	// Process each patient
	out := make(PatientMap, len(pm))
	var totalBefore, totalAfter int
	for eid, visits := range pm {
		cleanVisits := make([][]int, len(visits))
		for i, v := range visits {
			cleanVisits[i] = dedupeAndSort(v)
			totalBefore += len(cleanVisits[i])
		}
		newOnly := keepNewPerVisit(cleanVisits)
		for _, v := range newOnly {
			totalAfter += len(v)
		}
		out[eid] = newOnly
	}

	if err := savePatientMap(outPath, out); err != nil {
		fmt.Println("Error saving output:", err)
		os.Exit(1)
	}

}
