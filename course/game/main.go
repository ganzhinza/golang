package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
)

var rooms = map[string]*Room{}
var player Player

type Room struct {
	Name          string
	Location      string
	Locked        bool
	ways          map[string]*Room
	arrivalPhrase string
	things        map[string]bool
	clothes       map[string]bool
}

func (r Room) possibleWays() string {
	possibleWays := " можно пройти -"
	waysNames := make([]string, len(r.ways))
	i := 0
	for name := range r.ways {
		waysNames[i] = name
		i++
	}
	sort.Slice(waysNames, func(i, j int) bool {
		if r.ways[waysNames[i]].Location != r.ways[waysNames[j]].Location {
			return r.ways[waysNames[i]].Location < r.ways[waysNames[j]].Location
		} else {
			return waysNames[i] > waysNames[j]
		}
	})
	for _, name := range waysNames {
		possibleWays += " " + name + ","
	}
	return possibleWays[:len(possibleWays)-1]
}

func (r Room) arrive() string {
	return r.arrivalPhrase + r.possibleWays()
}

func (r Room) lookAround() string {
	res := ""
	switch r.Name {
	case "коридор":
		res = "ничего интересного. "
	case "улица":
		res = "на улице весна. "
	case "комната":
		roomThings := ""
		roomClothes := ""
		thingsList := make([]string, 0)
		clothesList := make([]string, 0)
		for key := range rooms["комната"].things {
			if rooms["комната"].things[key] {
				thingsList = append(thingsList, key)
			}
		}
		for key := range rooms["комната"].clothes {
			if rooms["комната"].clothes[key] {
				clothesList = append(clothesList, key)
			}
		}
		sort.Strings(thingsList)
		sort.Strings(clothesList)

		for _, val := range thingsList {
			roomThings += val + ", "
		}
		for _, val := range clothesList {
			roomClothes += val + ", "
		}

		if roomClothes == "" && roomThings == "" {
			res = "пустая комната."
		} else {
			roomThings = "на столе: " + roomThings
			roomClothes = "на стуле: " + roomClothes
			if roomThings != "на столе: " {
				res += roomThings
			}
			if roomClothes != "на стуле: " {
				res += roomClothes
			}
			res = res[:len(res)-2] + "."
		}
	case "кухня":
		if player.HasInventory && player.Inventory.things["конспекты"] && player.Inventory.things["ключи"] {
			res = "ты находишься на кухне, на столе: чай, надо идти в универ."
		} else {
			res = "ты находишься на кухне, на столе: чай, надо собрать рюкзак и идти в универ."
		}
	default:
		return "ничего интересного. "
	}
	return res + r.possibleWays()
}

type Inventory struct {
	Name   string
	things map[string]bool
}

type Player struct {
	HasInventory bool
	Inventory    Inventory
	CurrentRoom  *Room
}

func (p *Player) walk(roomName string) string {
	if val, ok := p.CurrentRoom.ways[roomName]; ok {
		if p.CurrentRoom.Location != val.Location && val.Locked {
			return "дверь закрыта"
		} else {
			p.CurrentRoom = val
		}
	} else {
		return "нет пути в " + roomName
	}
	return p.CurrentRoom.arrive()
}

func (p *Player) wear(inventory string) string {
	if p.CurrentRoom.clothes[inventory] {
		p.CurrentRoom.clothes[inventory] = false
		p.HasInventory = true
		return "вы надели: " + inventory
	}
	return "нет такого"
}

func (p *Player) take(thing string) string {
	if p.HasInventory {
		if p.CurrentRoom.things[thing] {
			p.CurrentRoom.things[thing] = false
			p.Inventory.things[thing] = true
			return "предмет добавлен в инвентарь: " + thing
		} else {
			return "нет такого"
		}
	} else {
		return "некуда класть"
	}
}

func (p *Player) use(thing, object string) string {
	if p.HasInventory && p.Inventory.things[thing] {
		if thing == "ключи" && object == "дверь" {
			for _, val := range p.CurrentRoom.ways {
				if val.Location != p.CurrentRoom.Location {
					val.Locked = false
					return "дверь открыта"
				}
			}
			return "не к чему применить"
		} else {
			return "не к чему применить"
		}
	} else {
		return "нет предмета в инвентаре - " + thing
	}
}

func main() {
	res := ""
	initGame()

	scanner := bufio.NewScanner(os.Stdin)

	for res != "выход" {
		scanner.Scan()
		res = scanner.Text()
		fmt.Printf("%s\n", handleCommand(res))
	}
}

func initGame() {
	rooms = make(map[string]*Room)
	rooms["кухня"] = &Room{Name: "кухня", arrivalPhrase: "кухня, ничего интересного."}
	rooms["коридор"] = &Room{Name: "коридор", arrivalPhrase: "ничего интересного."}
	rooms["комната"] = &Room{Name: "комната", arrivalPhrase: "ты в своей комнате."}
	rooms["улица"] = &Room{Name: "улица", arrivalPhrase: "на улице весна."}

	for _, val := range rooms {
		val.ways = make(map[string]*Room)
		val.things = make(map[string]bool)
		val.clothes = make(map[string]bool)
		val.Location = "дом"
	}
	rooms["улица"].Location = "улица"
	rooms["улица"].Locked = true

	rooms["комната"].things["ключи"] = true
	rooms["комната"].things["конспекты"] = true
	rooms["комната"].clothes["рюкзак"] = true

	rooms["кухня"].ways["коридор"] = rooms["коридор"]
	rooms["комната"].ways["коридор"] = rooms["коридор"]
	rooms["улица"].ways["домой"] = rooms["коридор"]

	rooms["коридор"].ways["комната"] = rooms["комната"]
	rooms["коридор"].ways["кухня"] = rooms["кухня"]
	rooms["коридор"].ways["улица"] = rooms["улица"]

	player.CurrentRoom = rooms["кухня"]
	player.HasInventory = false
	player.Inventory.things = make(map[string]bool)
}

func handleCommand(command string) string {
	commandSplitted := strings.Split(command, " ")
	switch commandSplitted[0] {
	case "идти":
		if len(commandSplitted) != 2 {
			return "неподходящее число аргументов"
		}
		return player.walk(commandSplitted[1])
	case "надеть":
		if len(commandSplitted) != 2 {
			return "неподходящее число аргументов"
		}
		return player.wear(commandSplitted[1])
	case "взять":
		if len(commandSplitted) != 2 {
			return "неподходящее число аргументов"
		}
		return player.take(commandSplitted[1])
	case "применить":
		if len(commandSplitted) != 3 {
			return "неподходящее число аргументов"
		}
		return player.use(commandSplitted[1], commandSplitted[2])
	case "осмотреться":
		if len(commandSplitted) != 1 {
			return "неподходящее число аргументов"
		}
		return player.CurrentRoom.lookAround()
	default:
		return "неизвестная команда"
	}
}
