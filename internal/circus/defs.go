// Package circus is the simulation behind Big Top Builder: the circus grounds,
// the tents and their shows, the performers, the guests and the money. It has
// no graphics, so it can be tested on its own.
package circus

// Cat is a building's category; it decides which build tab it's on.
type Cat int

const (
	CatTent Cat = iota
	CatSideshow
	CatFood
	CatExtra
)

var CatNames = [...]string{"Tents", "Sideshows", "Food", "Extras"}

// Def describes one kind of building.
type Def struct {
	ID, Name string
	Cat      Cat
	W, H     int // footprint in tiles; guests line up below the middle column
	Cost     int // to build
	Price    int // default ticket price
	MaxPrice int
	FairMax  int     // guests start to grumble above this price
	Capacity int     // guests per show, visit or serving
	Cycle    float64 // seconds per show, visit or serving
	Fun      float64 // happiness a sideshow gives (tents: set by the show)
	Hunger   float64 // how much it fills you up
	Thirst   float64
	Rest     float64 // energy back (benches)
	Upkeep   int     // running costs per night
	Goods    float64 // cost of the food per serving
	Slots    int     // tents: how many acts the show can have
	Tall     bool    // tents: high enough for the trapeze
	QueueMax int
	Light    float64 // radius it lights at night, in tiles
	Locked   bool    // shown but saved for the full game
	Blurb    string
}

// Visitable reports whether guests go to it (trash cans and lamps they don't).
func (d *Def) Visitable() bool { return d.Capacity > 0 }

var Defs = []*Def{
	// Tents
	{ID: "ringtent", Name: "Ring Tent", Cat: CatTent, W: 3, H: 3, Cost: 2200, Price: 4, MaxPrice: 12, FairMax: 5, Capacity: 14, Cycle: 22, Slots: 3, Upkeep: 120, QueueMax: 18, Light: 2.4,
		Blurb: "A one-ring tent. Book up to three acts and plan the running order."},
	{ID: "bigtop", Name: "The Big Top", Cat: CatTent, W: 4, H: 4, Cost: 5400, Price: 7, MaxPrice: 18, FairMax: 8, Capacity: 30, Cycle: 32, Slots: 5, Tall: true, Upkeep: 260, QueueMax: 34, Light: 3.6,
		Blurb: "The main event. Five acts, room for the trapeze, and the whole town talking."},
	{ID: "threering", Name: "Three-Ring Big Top", Cat: CatTent, W: 5, H: 4, Cost: 12000, Locked: true, Blurb: "Coming in the full game."},

	// Sideshows
	{ID: "funhouse", Name: "Funhouse", Cat: CatSideshow, W: 2, H: 2, Cost: 1100, Price: 2, MaxPrice: 7, FairMax: 3, Capacity: 6, Cycle: 7, Fun: 24, Upkeep: 50, QueueMax: 10, Light: 1.6,
		Blurb: "Wobbly floors, air jets and a giant clown mouth for a door."},
	{ID: "mirrors", Name: "Hall of Mirrors", Cat: CatSideshow, W: 2, H: 2, Cost: 850, Price: 2, MaxPrice: 6, FairMax: 3, Capacity: 5, Cycle: 6, Fun: 19, Upkeep: 35, QueueMax: 8, Light: 1.6,
		Blurb: "Tall, short, wide and wiggly. Guests come out giggling."},
	{ID: "fortune", Name: "Fortune Teller", Cat: CatSideshow, W: 2, H: 2, Cost: 600, Price: 3, MaxPrice: 8, FairMax: 4, Capacity: 1, Cycle: 3, Fun: 16, Upkeep: 40, QueueMax: 6, Light: 1.4,
		Blurb: "Madame Zora sees great things in your future. For a small fee."},
	{ID: "cannon", Name: "Human Cannonball", Cat: CatSideshow, W: 3, H: 2, Cost: 4200, Locked: true, Blurb: "Coming in the full game."},

	// Food
	{ID: "popcorn", Name: "Popcorn", Cat: CatFood, W: 2, H: 2, Cost: 480, Price: 3, MaxPrice: 8, FairMax: 4, Capacity: 1, Cycle: 1.8, Fun: 6, Hunger: 40, Thirst: -10, Upkeep: 40, Goods: 0.6, QueueMax: 6, Light: 1.6,
		Blurb: "Hot, buttery and salty. Makes guests thirsty."},
	{ID: "cottoncandy", Name: "Cotton Candy", Cat: CatFood, W: 2, H: 2, Cost: 520, Price: 3, MaxPrice: 8, FairMax: 4, Capacity: 1, Cycle: 2.0, Fun: 12, Hunger: 22, Upkeep: 40, Goods: 0.5, QueueMax: 6, Light: 1.6,
		Blurb: "Pink, fluffy and gone in a minute. Kids adore it."},
	{ID: "hotdogs", Name: "Hot Dogs", Cat: CatFood, W: 2, H: 2, Cost: 560, Price: 4, MaxPrice: 9, FairMax: 5, Capacity: 1, Cycle: 2.2, Fun: 5, Hunger: 62, Thirst: -6, Upkeep: 50, Goods: 1.2, QueueMax: 6, Light: 1.6,
		Blurb: "A proper meal between shows."},
	{ID: "soda", Name: "Soda Pop", Cat: CatFood, W: 2, H: 2, Cost: 380, Price: 2, MaxPrice: 6, FairMax: 3, Capacity: 1, Cycle: 1.6, Fun: 5, Thirst: 65, Upkeep: 35, Goods: 0.4, QueueMax: 6, Light: 1.6,
		Blurb: "Ice-cold sarsaparilla and orange fizz."},

	// Extras
	{ID: "bench", Name: "Bench", Cat: CatExtra, W: 1, H: 1, Cost: 80, Capacity: 2, Cycle: 7, Rest: 55, QueueMax: 2,
		Blurb: "Tired guests sit down and stay longer."},
	{ID: "trash", Name: "Trash Can", Cat: CatExtra, W: 1, H: 1, Cost: 60,
		Blurb: "Guests nearby don't drop litter."},
	{ID: "lamp", Name: "Light Pole", Cat: CatExtra, W: 1, H: 1, Cost: 140, Upkeep: 8, Light: 3.2,
		Blurb: "Dark corners make guests uneasy. Light them up."},
}

// DefByID finds a definition, or nil.
func DefByID(id string) *Def {
	for _, d := range Defs {
		if d.ID == id {
			return d
		}
	}
	return nil
}

// Style is the feeling an act gives the crowd. Shows work best with a mix.
type Style int

const (
	Laughs Style = iota
	Wonder
	Thrills
)

var StyleNames = [...]string{"Laughs", "Wonder", "Thrills"}

// ActDef describes a kind of performer.
type ActDef struct {
	ID, Name string
	Style    Style
	Appeal   float64 // how much the crowd loves a good one
	Hire     int     // signing fee
	Wage     int     // per night
	Tall     bool    // needs a tall tent (the Big Top)
	Blurb    string
}

var Acts = []*ActDef{
	{ID: "clown", Name: "Clown", Style: Laughs, Appeal: 20, Hire: 300, Wage: 45, Blurb: "Pratfalls, pies and a car that holds far too many friends."},
	{ID: "stilts", Name: "Stilt Walker", Style: Laughs, Appeal: 17, Hire: 250, Wage: 40, Blurb: "Ten feet tall and always waving."},
	{ID: "juggler", Name: "Juggler", Style: Wonder, Appeal: 23, Hire: 380, Wage: 55, Blurb: "Clubs, rings and the occasional flaming torch."},
	{ID: "strong", Name: "Strongwoman", Style: Wonder, Appeal: 27, Hire: 480, Wage: 65, Blurb: "Lifts a piano with a pianist still playing it."},
	{ID: "acrobats", Name: "Acrobats", Style: Thrills, Appeal: 31, Hire: 650, Wage: 85, Blurb: "Tumbling, pyramids and a backflip off a seesaw."},
	{ID: "fire", Name: "Fire Breather", Style: Thrills, Appeal: 34, Hire: 820, Wage: 100, Blurb: "Breathes a fireball over the front row. Gasps guaranteed."},
	{ID: "trapeze", Name: "Trapeze Artist", Style: Thrills, Appeal: 42, Hire: 1150, Wage: 140, Tall: true, Blurb: "Flies high under the canvas. Needs the Big Top."},
}

// ActByID finds an act, or nil.
func ActByID(id string) *ActDef {
	for _, a := range Acts {
		if a.ID == id {
			return a
		}
	}
	return nil
}

// Town is where the circus is set up. The demo has one; the full game goes on tour.
type Town struct {
	Name   string
	Nights int
	W, H   int
	Rep    float64 // the town's starting opinion of you
	Next   string  // where the wagons go after the demo
}

var DemoTown = Town{Name: "Brookhaven", Nights: 3, W: 24, H: 14, Rep: 2.5, Next: "Millbrook"}

const (
	StartCash    = 9000
	NightSeconds = 210.0 // 6 PM to 11 PM
	CleanupEach  = 2     // per piece of litter at closing
	DefaultGate  = 3
	MaxGate      = 12
	SellBack     = 0.6
	ShowEnergy   = 11.0 // energy an act spends per show
	TiredAt      = 35.0 // below this an act may flub
	ForHire      = 5    // performers looking for work each morning
	MaxTroupe    = 12
)
