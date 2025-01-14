package main

import (
	"bufio"
	"fmt"
	"github.com/FFX01/pokedex/internal/pokeApi"
	"math/rand"
	"os"
	"strings"
	"sync"
)

type Pokedex struct {
	mu      sync.Mutex
	Pokemon map[string]pokeApi.Pokemon
}

func newPokedex() *Pokedex {
	p := Pokedex{
		Pokemon: make(map[string]pokeApi.Pokemon),
	}
	return &p
}

func (p *Pokedex) list() {
	fmt.Println("Your Pokedex:")
	for _, pmon := range p.Pokemon {
		fmt.Printf("  - %s\n", pmon.Name)
	}
}

func (p *Pokedex) add(name string, details pokeApi.Pokemon) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Pokemon[name] = details
	return nil
}

func (p *Pokedex) get(name string) (pokeApi.Pokemon, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	details, ok := p.Pokemon[name]
	if !ok {
		return pokeApi.Pokemon{}, fmt.Errorf("You have not caught %s!; it is not in your pokedex.", name)
	}

	return details, nil
}

var pokedex *Pokedex = newPokedex()

func main() {
	stdin := bufio.NewScanner(os.Stdin)
	commandMap := getCommandMap()

	for {
		fmt.Printf("pokedex> ")
		stdin.Scan()
		userInput := stdin.Text()
		userCommand, args, err := parseUserInput(userInput)
		if err != nil {
			fmt.Println("Error: Invalid command")
		}
		com, ok := commandMap[userCommand]
		if !ok {
			fmt.Printf("Command: %s does not exist. Type \"help\" for usage information\n", userCommand)
			continue
		}
		err = com.callback(args)
		if err != nil {
			msg := fmt.Errorf("Error! Could not complete command: %w\n", err)
			fmt.Println(msg)
			continue
		}
	}
}

func parseUserInput(input string) (command string, args []string, error error) {
	items := strings.Split(input, " ")
	if len(items) > 2 {
		return "", []string{}, fmt.Errorf("Expected at most 2 arguments but got %v", items)
	}
	return items[0], items[1:], nil
}

type Command struct {
	name        string
	description string
	callback    func(args []string) error
}

func (c Command) print() error {
	fmt.Printf("%s: %s", c.name, c.description)
	return nil
}

func getCommandMap() map[string]Command {
	commands := map[string]Command{
		"help": {
			name:        "help",
			description: "Prints a help message",
			callback:    commandHelp,
		},
		"exit": {
			name:        "exit",
			description: "Exits the program",
			callback:    commandExit,
		},
		"map": {
			name:        "map",
			description: "Get a list of locations",
			callback:    commandMap,
		},
		"mapb": {
			name:        "mapb",
			description: "Get the last page of locations",
			callback:    commandMapB,
		},
		"explore": {
			name:        "explore",
			description: "List pokemon in the location",
			callback:    commandExplore,
		},
		"catch": {
			name:        "catch",
			description: "Catch a pokemon",
			callback:    CommandCatch,
		},
		"inspect": {
			name:        "inspect",
			description: "Inspect a pokemon in your pokedex",
			callback:    CommandInspect,
		},
		"pokedex": {
			name:        "pokedex",
			description: "List the pokemon in your Pokedex",
			callback:    CommandPokedex,
		},
	}
	return commands
}

func commandHelp(args []string) error {
	fmt.Println("Welcome to pokedex!")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println()

	commandMap := getCommandMap()
	for _, command := range commandMap {
		command.print()
		fmt.Print("\n")
	}

	return nil
}

func commandExit(args []string) error {
	fmt.Println("Now exiting pokedex...")
	os.Exit(0)
	return nil
}

func commandMap(args []string) error {
	results, err := pokeApi.GetMaps(false)
	if err != nil {
		return fmt.Errorf("Could not get locations: %w", err)
	}

	for _, area := range results.Results {
		fmt.Println(area.Name)
	}
	return nil
}

func commandMapB(args []string) error {
	results, err := pokeApi.GetMaps(true)
	if err != nil {
		return fmt.Errorf("Could not get locations: %w", err)
	}

	for _, area := range results.Results {
		fmt.Println(area.Name)
	}

	return nil
}

func commandExplore(args []string) error {
	results, err := pokeApi.GetPokemon(args[0])
	if err != nil {
		return fmt.Errorf("Could not get pokemon: %w", err)
	}

	for _, pokemon := range results.PokemonEncounters {
		fmt.Println(pokemon.Pokemon.Name)
	}
	return nil
}

func catch(experience int) bool {
	threshold := experience / 2
	roll := rand.Intn(experience)
	return roll >= threshold
}

func CommandCatch(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("You must pass the name of the pokemopn you want to catch")
	}
	pokemonName := args[0]

	fmt.Printf("Throwing a Pokeball at %s...\n", pokemonName)

	details, err := pokeApi.GetPokemonDetail(pokemonName)
	if err != nil {
		return fmt.Errorf("Error fetching details for %s: %w", pokemonName, err)
	}

	success := catch(details.BaseExperience)
	if success {
		pokedex.add(pokemonName, details)
		fmt.Printf("You cought %s!\n", pokemonName)
		fmt.Println("You may now inspect it with the `inspect` command")
	} else {
		fmt.Printf("You didn't catch %s!\n", pokemonName)
	}

	return nil
}

func CommandInspect(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("You must provide the name of the Pokemon you want to inspect")
	}
	pokemonName := args[0]

	details, err := pokedex.get(pokemonName)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	details.InspectPrint()

	return nil
}

func CommandPokedex(args []string) error {
	pokedex.list()
	return nil
}
