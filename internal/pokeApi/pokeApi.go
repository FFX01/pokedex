package pokeApi

import (
	"encoding/json"
	"fmt"
	"github.com/FFX01/pokedex/internal/pokecache"
	"io"
	"net/http"
	"time"
)

var config *Config = newConfig(time.Second * 10)

const baseUrl string = "https://pokeapi.co/api/v2/"

type Endpoint string

const (
	locationArea  Endpoint = "location-area"
	pokemonDetail Endpoint = "pokemon"
)

type Config struct {
	mapNext     *string
	mapPrevious *string
	cache       *pokecache.Cache
}

func newConfig(ttl time.Duration) *Config {
	cache := pokecache.NewCache(ttl)
	config := Config{
		cache: cache,
	}
	return &config
}

type Response interface {
	Print() error
}

type ListResponse[T Response] struct {
	Count    int
	Next     *string
	Previous *string
	Results  []T
}

func (lr ListResponse[T]) Print() error {
	repr, err := json.MarshalIndent(lr, "", "\t")
	if err != nil {
		return fmt.Errorf("Could not pretty print response: %w", err)
	}
	fmt.Print(string(repr) + "\n")
	return nil
}

type PokemonResponse[T Response] struct {
	PokemonEncounters []struct {
		Pokemon Pokemon
	} `json:"pokemon_encounters"`
}

type LocationArea struct {
	Name string
	Url  string
}

func (la LocationArea) Print() error {
	repr, err := json.MarshalIndent(la, "", "\t")
	if err != nil {
		return fmt.Errorf("Could not print response: %w", err)
	}
	fmt.Println(string(repr))
	return nil
}

type Pokemon struct {
	Id             int
	Name           string
	Url            string
	BaseExperience int `json:"base_experience"`
	Height         int
	Weight         int
	Abilities      []struct {
		Ability struct {
			Name string
		}
	}
	Moves []struct {
		Move struct {
			Name string
		}
	}
	Stats []struct {
		BaseStat int `json:"base_stat"`
		Stat     struct {
			Name string
		}
	}
	Types []struct {
		Type struct {
			Name string
		}
	}
}

func (p Pokemon) InspectPrint() {
	fmt.Printf("Name: %s\n", p.Name)
	fmt.Printf("Base Experience: %v\n", p.BaseExperience)
	fmt.Printf("Height: %v\n", p.Height)
	fmt.Printf("Weight: %v\n", p.Weight)

    fmt.Println("Types:")
    for _, t := range p.Types {
        fmt.Printf("  - %s\n", t.Type.Name)
    }

	fmt.Println("Stats:")
	for _, stat := range p.Stats {
		fmt.Printf("  - %s: %v\n", stat.Stat.Name, stat.BaseStat)
	}

    fmt.Println("Moves:")
    for _, m := range p.Moves {
        fmt.Printf("  - %s\n", m.Move.Name)
    }

    fmt.Println("Abilities:")
    for _, a := range p.Abilities {
        fmt.Printf("  - %s\n", a.Ability.Name)
    }
}

func (p Pokemon) Print() error {
	repr, err := json.MarshalIndent(p, "", "\t")
	if err != nil {
		return fmt.Errorf("Could not print pokemon response:  %w", err)
	}
	fmt.Println(string(repr))
	return nil
}

func GetMaps(back bool) (ListResponse[LocationArea], error) {
	url := baseUrl
	if !back {
		if config.mapNext == nil {
			url += string(locationArea)
		} else {
			url = *config.mapNext
		}
	} else {
		if config.mapPrevious == nil {
			return ListResponse[LocationArea]{}, fmt.Errorf("There is no previous page")
		} else {
			url = *config.mapPrevious
		}
	}

	resp, err := getRequest(url)
	if err != nil {
		return ListResponse[LocationArea]{}, fmt.Errorf("Error getting maps: %w", err)
	}

	var data ListResponse[LocationArea]
	err = json.Unmarshal(resp, &data)
	if err != nil {
		return data, fmt.Errorf("Error parsing json: %w", err)
	}
	config.mapNext = data.Next
	config.mapPrevious = data.Previous

	return data, nil
}

func GetPokemon(location string) (PokemonResponse[Pokemon], error) {
	url := baseUrl + string(locationArea) + "/" + location + "/"

	resp, err := getRequest(url)
	if err != nil {
		return PokemonResponse[Pokemon]{}, fmt.Errorf("Error getting pokemon: %w", err)
	}

	var data PokemonResponse[Pokemon]
	err = json.Unmarshal(resp, &data)
	if err != nil {
		return PokemonResponse[Pokemon]{}, fmt.Errorf("Could not parse pokemon response: %w", err)
	}

	return data, nil
}

func GetPokemonDetail(name string) (Pokemon, error) {
	url := baseUrl + string(pokemonDetail) + "/" + name

	resp, err := getRequest(url)
	if err != nil {
		return Pokemon{}, fmt.Errorf("Error fetching pokemon details: %w", err)
	}

	var data Pokemon
	err = json.Unmarshal(resp, &data)
	if err != nil {
		return Pokemon{}, fmt.Errorf("Could not parse response: %w", err)
	}

	return data, nil
}

func getRequest(url string) ([]byte, error) {
	// Check cache for url
	cachedData, found := config.cache.Get(url)
	if found {
		return cachedData, nil
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return []byte{}, fmt.Errorf("Error creating request: %w", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return []byte{}, fmt.Errorf("Error sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []byte{}, fmt.Errorf("Request Error(%s): status %d, %w", url, resp.StatusCode, err)
	}

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, fmt.Errorf("Error reading response body: %w", err)
	}
	config.cache.Add(url, bytes)
	return bytes, nil
}
