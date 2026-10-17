package data

// Seed catalog. All titles, people, chains and venues are fictional; posters
// and banners are generated procedurally by the catalog image service.

type Person struct {
	Name string `json:"name"`
	Role string `json:"role"`
}

type Movie struct {
	Code        string
	Title       string
	Tagline     string
	Languages   []string
	Formats     []string
	Genres      []string
	Certificate string
	DurationMin int
	ReleaseDays int // release date relative to seed day; positive = coming soon
	Synopsis    string
	Cast        []Person
	Crew        []Person
	Rating      float64 // 0 = not rated yet (show likes instead)
	Votes       int
	Likes       int
	Promoted    bool
	Hue         int
	Popularity  int
}

func (m Movie) NowShowing() bool { return m.ReleaseDays <= 0 }

var Movies = []Movie{
	{Code: "ET00510001", Title: "Tidebreaker", Tagline: "The ocean keeps its secrets. Until now.", Languages: []string{"Hindi", "Tamil", "Telugu"}, Formats: []string{"2D", "IMAX 2D", "4DX"}, Genres: []string{"Action", "Thriller"}, Certificate: "UA16+", DurationMin: 158, ReleaseDays: -9,
		Synopsis: "A disgraced naval diver is pulled back for one last salvage run off the Konkan coast, only to discover that the sunken freighter hides a conspiracy reaching the highest offices in Delhi.",
		Cast:     []Person{{"Arjun Vasant", "as Commander Rudra"}, {"Meher Kalra", "as Dr. Ira Sen"}, {"Devraj Iyer", "as Kailash"}, {"Tanvi Rao", "as Neelam"}, {"Kabir Malhotra", "as Shekhar"}},
		Crew:     []Person{{"Rohan Dutt", "Director"}, {"Sana Qureshi", "Producer"}, {"Aditya Menon", "Musician"}, {"Farhan Lakdawala", "Cinematographer"}},
		Rating:   8.6, Votes: 128400, Promoted: true, Hue: 200, Popularity: 98},
	{Code: "ET00510002", Title: "Saffron Skies", Tagline: "Love lands where it wants.", Languages: []string{"Hindi"}, Formats: []string{"2D"}, Genres: []string{"Romantic", "Drama"}, Certificate: "UA", DurationMin: 142, ReleaseDays: -2,
		Synopsis: "A small-town paragliding instructor in Bir Billing and a burnt-out Mumbai architect find their carefully planned lives drifting off course over one unforgettable monsoon season.",
		Cast:     []Person{{"Ishaan Grover", "as Kabir"}, {"Riya Deshmukh", "as Naina"}, {"Pankaj Saxena", "as Bauji"}, {"Lata Venkatesh", "as Mrs. D'Souza"}},
		Crew:     []Person{{"Nandita Bose", "Director"}, {"Vikram Chandra Rao", "Producer"}, {"Shreya Iyengar", "Musician"}},
		Rating:   8.1, Votes: 21700, Promoted: true, Hue: 28, Popularity: 91},
	{Code: "ET00510003", Title: "Kaal Bhairav", Tagline: "Time bows to no one. Except him.", Languages: []string{"Telugu", "Hindi", "Tamil", "Kannada", "Malayalam"}, Formats: []string{"2D", "3D", "IMAX 3D"}, Genres: []string{"Action", "Fantasy", "Period"}, Certificate: "UA13+", DurationMin: 171, ReleaseDays: -16,
		Synopsis: "In a forgotten Deccan kingdom, a temple guardian bound to an ancient vow must rise against an empire that has learned to bend time itself.",
		Cast:     []Person{{"Aditya Varma Raju", "as Bhairav"}, {"Sravani Devi", "as Rani Mrinalini"}, {"Jagannath Rao", "as Emperor Vikramaditya"}, {"Venkatesh Prasad", "as Guru Shankaracharya"}, {"Divya Pillai", "as Chandrika"}},
		Crew:     []Person{{"K. Mahendra", "Director"}, {"Charminar Creations", "Producer"}, {"Sai Charan", "Musician"}},
		Rating:   9.1, Votes: 342000, Hue: 12, Popularity: 97},
	{Code: "ET00510004", Title: "The Quiet Protocol", Tagline: "Some codes should never be broken.", Languages: []string{"English", "Hindi"}, Formats: []string{"2D", "IMAX 2D", "DOLBY CINEMA 2D"}, Genres: []string{"Sci-Fi", "Thriller"}, Certificate: "UA13+", DurationMin: 136, ReleaseDays: -23,
		Synopsis: "When an AI built to negotiate peace treaties goes silent mid-summit, a linguist and a hacker have forty-eight hours to understand what it heard.",
		Cast:     []Person{{"Clara Whitfield", "as Dr. Cass Holloway"}, {"Neel Bhasin", "as Raghav"}, {"Marcus Lane", "as Director Pryce"}, {"Yuki Tanaka", "as Hana"}},
		Crew:     []Person{{"Owen Castellanos", "Director"}, {"Northlight Pictures", "Producer"}, {"Calder Brooks", "Musician"}},
		Rating:   8.3, Votes: 54200, Hue: 230, Popularity: 88},
	{Code: "ET00510005", Title: "Maska Bun", Tagline: "One café. Three generations. Zero chill.", Languages: []string{"Hindi", "Marathi"}, Formats: []string{"2D"}, Genres: []string{"Comedy", "Family"}, Certificate: "U", DurationMin: 128, ReleaseDays: -5,
		Synopsis: "The squabbling Irani family must save their hundred-year-old Pune café from a slick franchise chain, armed only with bun maska, chai and terrible ideas.",
		Cast:     []Person{{"Cyrus Bharucha", "as Rustom"}, {"Shirin Mistry", "as Dilnaz"}, {"Zubin Engineer", "as Jimmy"}, {"Homi Wadia", "as Grandpa Feroze"}},
		Crew:     []Person{{"Kiran Joshi", "Director"}, {"Chai Pani Films", "Producer"}, {"Kunal Ganjawala Sane", "Musician"}},
		Rating:   8.8, Votes: 16900, Hue: 45, Popularity: 84},
	{Code: "ET00510006", Title: "Velvet Static", Tagline: "Tune in. Freak out.", Languages: []string{"English"}, Formats: []string{"2D", "4DX"}, Genres: []string{"Horror", "Mystery"}, Certificate: "A", DurationMin: 109, ReleaseDays: -12,
		Synopsis: "A late-night radio host starts receiving calls from listeners who died decades ago — and every caller knows something about her that she has never told anyone.",
		Cast:     []Person{{"Harper Quinn", "as June Avery"}, {"Tomas Velez", "as Eli"}, {"Grace Okafor", "as Detective Morrow"}},
		Crew:     []Person{{"Lena Strand", "Director"}, {"Blackwire Films", "Producer"}},
		Rating:   7.4, Votes: 8800, Hue: 290, Popularity: 72},
	{Code: "ET00510007", Title: "Thalaivan Returns", Tagline: "The boss is back.", Languages: []string{"Tamil", "Telugu", "Hindi"}, Formats: []string{"2D", "IMAX 2D"}, Genres: []string{"Action", "Drama"}, Certificate: "UA16+", DurationMin: 165, ReleaseDays: -1,
		Synopsis: "Twenty years after vanishing from Madurai, a feared don returns as a mild-mannered school principal — until his past arrives at the school gate.",
		Cast:     []Person{{"Senthil Arumugam", "as Thalaivan"}, {"Kavya Lakshmanan", "as Meenakshi"}, {"Prabhu Ganesan", "as ACP Karthik"}, {"Muthu Pandian", "as Peon Murugan"}},
		Crew:     []Person{{"K. Selvaraghavan Mani", "Director"}, {"Kovil Talkies", "Producer"}, {"Ilango Vel", "Musician"}},
		Likes:    183000, Hue: 0, Popularity: 99},
	{Code: "ET00510008", Title: "Tiny Titans: Dhol Ka Raaz", Tagline: "Dhunpur's biggest mystery!", Languages: []string{"Hindi", "English", "Tamil", "Telugu"}, Formats: []string{"2D", "3D"}, Genres: []string{"Animation", "Adventure", "Family"}, Certificate: "U", DurationMin: 104, ReleaseDays: -19,
		Synopsis: "When the royal drum that protects Dhunpur goes missing on the eve of the harvest festival, a brave boy and his friends follow a trail of riddles into the Whispering Hills.",
		Cast:     []Person{{"Pihu Mathur", "as Veer (voice)"}, {"Neha Sahni", "as Gudiya (voice)"}, {"Trisha Bhalla", "as Raju (voice)"}},
		Crew:     []Person{{"Sameer Kotian", "Director"}, {"Pebble Toons", "Producer"}},
		Rating:   8.9, Votes: 6100, Hue: 120, Popularity: 70},
	{Code: "ET00510009", Title: "Aranyam", Tagline: "The forest remembers.", Languages: []string{"Malayalam", "Tamil"}, Formats: []string{"2D"}, Genres: []string{"Thriller", "Mystery"}, Certificate: "UA16+", DurationMin: 141, ReleaseDays: -30,
		Synopsis: "A forest officer investigating a string of disappearances in Wayanad uncovers a tribal legend that the timber mafia would kill to keep buried.",
		Cast:     []Person{{"Rahul Thomas", "as Ranger Joseph"}, {"Anju Varghese", "as Dr. Anitha"}, {"Biju Kurian", "as Kuttan"}},
		Crew:     []Person{{"Jeo Mathew", "Director"}, {"Kayal Films", "Producer"}, {"Rex Joseph", "Musician"}},
		Rating:   9.2, Votes: 45600, Hue: 140, Popularity: 82},
	{Code: "ET00510010", Title: "Seoul Mates", Tagline: "Two strangers. One apartment. No common language.", Languages: []string{"Korean", "Hindi"}, Formats: []string{"2D"}, Genres: []string{"Romantic", "Comedy"}, Certificate: "UA", DurationMin: 118, ReleaseDays: -7,
		Synopsis: "A K-drama-obsessed Delhi software engineer and a grumpy Seoul chef are accidentally rented the same apartment for a month — and neither will move out.",
		Cast:     []Person{{"Park Min-jun", "as Ji-ho"}, {"Aditi Chawla", "as Simran"}, {"Kim Soo-ah", "as Mrs. Lee"}},
		Crew:     []Person{{"Lee Hae-won", "Director"}, {"Han River Pictures", "Producer"}},
		Rating:   8.0, Votes: 4300, Hue: 330, Popularity: 66},
	{Code: "ET00510011", Title: "Ghoomta Garba", Tagline: "The dance of defiance.", Languages: []string{"Gujarati", "Hindi"}, Formats: []string{"2D"}, Genres: []string{"Drama", "Musical"}, Certificate: "U", DurationMin: 132, ReleaseDays: -14,
		Synopsis: "In a Kutch village where the garba circle has always been led by the headman's family, a widowed tailor decides this Navratri the women will dance to their own beat.",
		Cast:     []Person{{"Hetal Trivedi", "as Kesar"}, {"Chirag Vora", "as Hasmukh"}, {"Sudha Pandya", "as Ba"}},
		Crew:     []Person{{"Jignesh Bhatt", "Director"}, {"Rann Utsav Films", "Producer"}, {"Parth Dave", "Musician"}},
		Rating:   8.7, Votes: 3900, Hue: 330, Popularity: 61},
	{Code: "ET00510012", Title: "Pather Kotha", Tagline: "Every road tells a story.", Languages: []string{"Bengali"}, Formats: []string{"2D"}, Genres: []string{"Drama"}, Certificate: "U", DurationMin: 126, ReleaseDays: -21,
		Synopsis: "An ageing Kolkata tram conductor on his final week of service records the stories of his passengers, unaware that one of them is the son he never knew.",
		Cast:     []Person{{"Haridas Mitra", "as Haradhan"}, {"Sohini Dutta", "as Malini"}, {"Anirban Pal", "as Arko"}},
		Crew:     []Person{{"Subhro Chakraborty", "Director"}, {"Howrah Bridge Films", "Producer"}},
		Rating:   9.0, Votes: 7100, Hue: 35, Popularity: 58},
	{Code: "ET00510013", Title: "Iron Monsoon", Tagline: "When the storm hits, they rise.", Languages: []string{"English", "Hindi", "Tamil", "Telugu"}, Formats: []string{"2D", "3D", "IMAX 3D", "4DX"}, Genres: []string{"Action", "Sci-Fi", "Adventure"}, Certificate: "UA13+", DurationMin: 149, ReleaseDays: -4,
		Synopsis: "A crew of mech pilots from five nations must hold the Mumbai sea wall when a hurricane carries something far worse than rain onto the city's shores.",
		Cast:     []Person{{"Jack Hollander", "as Captain Rhys"}, {"Asmita Varma", "as Lt. Asha Varma"}, {"Kenji Moriyama", "as Dr. Mori"}, {"Lucia Ferraro", "as Gia"}},
		Crew:     []Person{{"Gareth Holm", "Director"}, {"Titan Frame", "Producer"}, {"Lars Ekholm", "Musician"}},
		Rating:   7.9, Votes: 96500, Promoted: true, Hue: 210, Popularity: 95},
	{Code: "ET00510014", Title: "Baap Re Baap", Tagline: "Two fathers. One wedding. Total chaos.", Languages: []string{"Hindi"}, Formats: []string{"2D"}, Genres: []string{"Comedy"}, Certificate: "UA", DurationMin: 137, ReleaseDays: -10,
		Synopsis: "A Punjabi caterer and a Tamil Brahmin professor go to war over every detail of their children's wedding, until a tax raid forces them to become partners in crime.",
		Cast:     []Person{{"Gurpreet Sandhu", "as Balwinder"}, {"Venkat Subramaniam", "as Prof. Iyer"}, {"Simran Ahluwalia", "as Pinky"}, {"Karthik Raman", "as Karthik"}},
		Crew:     []Person{{"Nitin Chawla", "Director"}, {"Dilli Talkies", "Producer"}},
		Likes:    54200, Hue: 15, Popularity: 80},
	{Code: "ET00510015", Title: "Daiva Kshetra", Tagline: "Faith has a fierce face.", Languages: []string{"Kannada", "Hindi", "Telugu"}, Formats: []string{"2D", "IMAX 2D"}, Genres: []string{"Action", "Drama", "Mythology"}, Certificate: "UA16+", DurationMin: 152, ReleaseDays: -6,
		Synopsis: "On the Tulu coast, a Bhoota Kola performer is caught between a land-grabbing politician and the spirit he is sworn to embody.",
		Cast:     []Person{{"Shravan Poojary", "as Shiva"}, {"Megha Shetty", "as Leela"}, {"Ramesh Bangera", "as Devendra"}},
		Crew:     []Person{{"Shravan Poojary", "Director"}, {"Karavali Pictures", "Producer"}, {"Vasuki Rao", "Musician"}},
		Rating:   9.3, Votes: 88100, Hue: 25, Popularity: 93},
	{Code: "ET00510016", Title: "Pune 411", Tagline: "Every lane has a secret.", Languages: []string{"Marathi"}, Formats: []string{"2D"}, Genres: []string{"Crime", "Thriller"}, Certificate: "A", DurationMin: 131, ReleaseDays: -3,
		Synopsis: "A rookie crime reporter and a retired inspector trace a decades-old serial case through the wadas and pethas of old Pune.",
		Cast:     []Person{{"Omkar Pendse", "as Aditya"}, {"Sayali Gokhale", "as Insp. Godbole"}, {"Madhav Apte", "as Bhau"}},
		Crew:     []Person{{"Sachin Bhosale", "Director"}, {"Shaniwar Wada Films", "Producer"}},
		Rating:   8.5, Votes: 5200, Hue: 260, Popularity: 63},
	{Code: "ET00510017", Title: "Gunda Raj", Tagline: "Rule. Or be ruled.", Languages: []string{"Telugu", "Hindi"}, Formats: []string{"2D", "4DX"}, Genres: []string{"Action", "Crime"}, Certificate: "A", DurationMin: 160, ReleaseDays: -26,
		Synopsis: "A Rayalaseema farmer's son climbs Hyderabad's underworld to avenge his father, only to find himself becoming the man he swore to destroy.",
		Cast:     []Person{{"Vamsi Krishna", "as Shiva Reddy"}, {"Lavanya Chowdary", "as Swathi"}, {"Satyanarayana Murthy", "as Narasimha"}},
		Crew:     []Person{{"Koti Srinivas", "Director"}, {"Charminar Creations", "Producer"}, {"Raghu Kunche", "Musician"}},
		Rating:   7.8, Votes: 61300, Hue: 355, Popularity: 76},
	{Code: "ET00510018", Title: "Little Lighthouse", Tagline: "Some lights guide you home.", Languages: []string{"English"}, Formats: []string{"2D", "3D"}, Genres: []string{"Animation", "Family", "Adventure"}, Certificate: "U", DurationMin: 97, ReleaseDays: -11,
		Synopsis: "A lonely lighthouse that dreams of seeing the world gets its wish when a storm sweeps it out to sea with a seagull, a crab and a very confused fisherman aboard.",
		Cast:     []Person{{"Emma Thornton", "as Lumi (voice)"}, {"Riz Patel", "as Gully (voice)"}, {"Oliver Grant", "as Captain Bo (voice)"}},
		Crew:     []Person{{"Pete Dockery", "Director"}, {"Brightwave Animation", "Producer"}},
		Rating:   8.4, Votes: 12800, Hue: 185, Popularity: 68},

	// Coming soon
	{Code: "ET00510019", Title: "Shadow of Durga", Tagline: "This Dussehra, evil has a new name.", Languages: []string{"Hindi", "Bengali"}, Formats: []string{"2D", "IMAX 2D"}, Genres: []string{"Action", "Mythology"}, Certificate: "UA16+", DurationMin: 155, ReleaseDays: 6,
		Synopsis: "A Kolkata idol-maker discovers that the clay she shapes each Durga Puja carries a power that an ancient cult has hunted for centuries.",
		Cast:     []Person{{"Ishita Majumdar", "as Uma"}, {"Arindam Bose", "as Mahishasur"}},
		Crew:     []Person{{"Dibakar Sen", "Director"}},
		Likes:    88200, Hue: 5, Popularity: 90},
	{Code: "ET00510020", Title: "Orbit 9", Tagline: "Mission control has stopped answering.", Languages: []string{"English", "Hindi"}, Formats: []string{"2D", "IMAX 2D", "DOLBY CINEMA 2D"}, Genres: []string{"Sci-Fi", "Drama"}, Certificate: "UA13+", DurationMin: 144, ReleaseDays: 13,
		Synopsis: "The first crewed Indian mission to Mars loses contact with Earth, and the astronauts must decide whether the silence is an accident or a warning.",
		Cast:     []Person{{"Sridhar Natarajan", "as Cdr. Narayan"}, {"Harleen Kaur", "as Dr. Kavya"}},
		Crew:     []Person{{"Anand Kulkarni", "Director"}},
		Likes:    41700, Hue: 250, Popularity: 85},
	{Code: "ET00510021", Title: "Rangeela Raja", Tagline: "Colour your heart.", Languages: []string{"Hindi"}, Formats: []string{"2D"}, Genres: []string{"Romantic", "Musical"}, Certificate: "U", DurationMin: 140, ReleaseDays: 20,
		Synopsis: "A Jaipur street magician falls for a Bollywood choreographer, and agrees to fake a royal lineage to win her family over.",
		Cast:     []Person{{"Yash Bhardwaj", "as Raja"}, {"Mira Kapadia", "as Mira"}},
		Crew:     []Person{{"Farid Siddiqui", "Director"}},
		Likes:    23900, Hue: 320, Popularity: 74},
	{Code: "ET00510022", Title: "Kaval Kadhai 2", Tagline: "Every story has two sides.", Languages: []string{"Tamil", "Hindi"}, Formats: []string{"2D", "IMAX 2D"}, Genres: []string{"Action", "Thriller"}, Certificate: "UA16+", DurationMin: 158, ReleaseDays: 27,
		Synopsis: "The inspector and the smuggler return, but this time the riddle they must solve together could bring down the entire Chennai police force.",
		Cast:     []Person{{"Arun Shanmugam", "as Inspector Arun"}, {"Durai Pandi", "as Durai"}},
		Crew:     []Person{{"Mani Chandrasekar", "Director"}},
		Likes:    67400, Hue: 30, Popularity: 87},
}

type Tier struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	Price    int    `json:"price"`
	Capacity int    `json:"capacity"`
	Note     string `json:"note,omitempty"`
}

type Event struct {
	Code        string
	Title       string
	Category    string // EVENTS, PLAYS, SPORTS, ACTIVITIES
	Genre       string
	Languages   []string
	AgeLimit    string
	DurationMin int
	Synopsis    string
	Region      string // region code; "" = every popular city (touring show)
	VenueName   string
	DayOffset   int
	Hour, Min   int
	Tiers       []Tier
	Promoted    bool
	Hue         int
	Popularity  int
}

var Events = []Event{
	{Code: "ET00520001", Title: "Ravi Kapoor: Unfiltered Tour", Category: "EVENTS", Genre: "Comedy Shows", Languages: []string{"Hindi"}, AgeLimit: "16yrs +", DurationMin: 90,
		Synopsis:  "A storyteller-comic is back on the road with a brand new hour of stories about family WhatsApp groups, arranged marriages and the great Indian middle class.",
		VenueName: "Laughter Arena", DayOffset: 2, Hour: 20, Tiers: []Tier{{"GA", "Silver", 799, 300, ""}, {"GD", "Gold", 1299, 150, "Rows A-F"}, {"VP", "Platinum", 2499, 40, "Front row + meet & greet"}}, Promoted: true, Hue: 40, Popularity: 96},
	{Code: "ET00520002", Title: "Bassline Arena ft. Neon Drift", Category: "EVENTS", Genre: "Music Shows", Languages: []string{"English"}, AgeLimit: "18yrs +", DurationMin: 300,
		Synopsis: "The biggest names in electronic music take over the city for one night of lasers, bass and fireworks. Doors open at 4 PM.",
		Region:   "MUMBAI", VenueName: "Bandra Reclamation Grounds", DayOffset: 8, Hour: 16, Tiers: []Tier{{"PH", "Phase 1 GA", 1999, 2000, ""}, {"VP", "VIP", 4999, 400, "Elevated deck, fast-track entry"}}, Promoted: true, Hue: 280, Popularity: 94},
	{Code: "ET00520003", Title: "Raag Utsav: An Evening of Carnatic Masters", Category: "EVENTS", Genre: "Music Shows", Languages: []string{"Tamil"}, AgeLimit: "All ages", DurationMin: 150,
		Synopsis: "A rare evening featuring three generations of Carnatic vocalists, with mridangam and violin accompaniment.",
		Region:   "CHEN", VenueName: "Sangeetha Sabha Hall", DayOffset: 5, Hour: 18, Min: 30, Tiers: []Tier{{"BL", "Balcony", 500, 300, ""}, {"ST", "Stalls", 1200, 200, ""}}, Hue: 30, Popularity: 60},
	{Code: "ET00520004", Title: "Hamlet in Hinglish", Category: "PLAYS", Genre: "Theatre", Languages: []string{"Hindi", "English"}, AgeLimit: "12yrs +", DurationMin: 140,
		Synopsis:  "Shakespeare's prince of Denmark reimagined as the heir to a Delhi construction empire. Sharp, funny and devastating.",
		VenueName: "Black Box Stage", DayOffset: 3, Hour: 19, Tiers: []Tier{{"ST", "Standard", 600, 180, ""}, {"PR", "Premium", 1000, 60, "Rows A-C"}}, Hue: 0, Popularity: 72},
	{Code: "ET00520005", Title: "Letters to Amrita", Category: "PLAYS", Genre: "Drama", Languages: []string{"Hindi", "Urdu"}, AgeLimit: "15yrs +", DurationMin: 100,
		Synopsis:  "An epistolary play: two friends, fifty years of letters and one unforgettable reading.",
		VenueName: "Rangmanch Auditorium", DayOffset: 6, Hour: 19, Min: 30, Tiers: []Tier{{"ST", "Standard", 499, 250, ""}, {"PR", "Premium", 899, 80, ""}}, Hue: 340, Popularity: 64},
	{Code: "ET00520006", Title: "Mumbai Mavericks vs Delhi Dynamos", Category: "SPORTS", Genre: "Cricket", Languages: []string{"English", "Hindi"}, AgeLimit: "All ages", DurationMin: 210,
		Synopsis: "League stage clash of the Premier T20 Cup. Gates open two hours before the first ball.",
		Region:   "MUMBAI", VenueName: "Marine Lines Stadium", DayOffset: 4, Hour: 19, Min: 30, Tiers: []Tier{{"NS", "North Stand", 1200, 5000, ""}, {"PV", "Pavilion", 3500, 1500, ""}, {"CB", "Corporate Box", 12000, 100, "Food & beverages included"}}, Promoted: true, Hue: 215, Popularity: 92},
	{Code: "ET00520007", Title: "Bengaluru Blasters vs Chennai Chargers", Category: "SPORTS", Genre: "Kabaddi", Languages: []string{"English", "Kannada"}, AgeLimit: "All ages", DurationMin: 120,
		Synopsis: "Pro Kabaddi action live. Feel the raid from courtside.",
		Region:   "BANG", VenueName: "Garden City Indoor Arena", DayOffset: 3, Hour: 20, Tiers: []Tier{{"GA", "General", 400, 2000, ""}, {"CS", "Courtside", 2500, 120, ""}}, Hue: 10, Popularity: 70},
	{Code: "ET00520008", Title: "City Half Marathon 2026", Category: "SPORTS", Genre: "Running", Languages: []string{"English"}, AgeLimit: "16yrs +", DurationMin: 240,
		Synopsis:  "21.1 km, 10 km and 5 km categories. Timing chip, finisher medal and breakfast included.",
		VenueName: "City Sports Complex", DayOffset: 9, Hour: 5, Min: 30, Tiers: []Tier{{"5K", "5K Fun Run", 799, 3000, ""}, {"10K", "10K", 1299, 2500, ""}, {"HM", "Half Marathon", 1899, 2000, ""}}, Hue: 160, Popularity: 66},
	{Code: "ET00520009", Title: "Splash Kingdom Day Pass", Category: "ACTIVITIES", Genre: "Amusement Park", Languages: []string{"English"}, AgeLimit: "All ages", DurationMin: 480,
		Synopsis:  "Sixty rides, a wave pool and the tallest water slide in the country. Skip-the-queue passes available.",
		VenueName: "Splash Kingdom", DayOffset: 1, Hour: 10, Tiers: []Tier{{"ST", "Standard", 1199, 2000, ""}, {"FQ", "Fast Queue", 1999, 300, "Priority access on 20 rides"}}, Hue: 195, Popularity: 78},
	{Code: "ET00520010", Title: "Pottery & Prosecco Workshop", Category: "ACTIVITIES", Genre: "Workshops", Languages: []string{"English", "Hindi"}, AgeLimit: "18yrs +", DurationMin: 150,
		Synopsis:  "Get your hands dirty on the wheel with a master potter. All materials included; your creation is fired and delivered in two weeks.",
		VenueName: "The Clay Studio", DayOffset: 2, Hour: 16, Tiers: []Tier{{"ST", "Single", 1800, 24, ""}, {"DU", "Duo", 3200, 12, "For two"}}, Hue: 25, Popularity: 55},
	{Code: "ET00520011", Title: "Kids' Science Carnival", Category: "ACTIVITIES", Genre: "Kids", Languages: []string{"English"}, AgeLimit: "4yrs +", DurationMin: 240,
		Synopsis:  "Slime labs, rocket launches and a walk-through planetarium for curious young minds.",
		VenueName: "Expo Centre Hall 3", DayOffset: 5, Hour: 11, Tiers: []Tier{{"KD", "Kid + 1 Adult", 699, 800, ""}}, Hue: 100, Popularity: 58},
	{Code: "ET00520012", Title: "Trek to Kalsubai Peak", Category: "ACTIVITIES", Genre: "Adventure", Languages: []string{"English", "Marathi"}, AgeLimit: "12yrs +", DurationMin: 720,
		Synopsis: "A guided monsoon trek to Maharashtra's highest summit. Transport from the city, breakfast and a certified trek leader included.",
		Region:   "PUNE", VenueName: "Bari Village Base Camp", DayOffset: 7, Hour: 5, Tiers: []Tier{{"ST", "With transport", 1499, 60, ""}}, Hue: 130, Popularity: 50},
	{Code: "ET00520013", Title: "Bollywood Retro Night", Category: "EVENTS", Genre: "Music Shows", Languages: []string{"Hindi"}, AgeLimit: "18yrs +", DurationMin: 180,
		Synopsis:  "A live 12-piece band plays the biggest hits from the 70s to the 90s. Dress code: as filmy as you dare.",
		VenueName: "The Grand Ballroom", DayOffset: 4, Hour: 21, Tiers: []Tier{{"GA", "Entry", 999, 400, ""}, {"TB", "Table for 4", 6999, 30, "Includes ₹4000 cover"}}, Hue: 300, Popularity: 75},
	{Code: "ET00520014", Title: "Open Mic Fridays", Category: "EVENTS", Genre: "Comedy Shows", Languages: []string{"Hindi", "English"}, AgeLimit: "18yrs +", DurationMin: 120,
		Synopsis:  "Ten new comics, ten minutes each. Some will be great. Some will be very, very brave.",
		VenueName: "The Habitat Café", DayOffset: 1, Hour: 21, Tiers: []Tier{{"GA", "Entry", 299, 120, ""}}, Hue: 50, Popularity: 57},
	{Code: "ET00520015", Title: "Dandiya Dhamaka", Category: "EVENTS", Genre: "Festivals", Languages: []string{"Gujarati", "Hindi"}, AgeLimit: "All ages", DurationMin: 300,
		Synopsis: "Nine nights of garba and dandiya with live folk singers, food stalls and best-dressed contests.",
		Region:   "AHD", VenueName: "Riverfront Grounds", DayOffset: 1, Hour: 20, Tiers: []Tier{{"DP", "Day Pass", 499, 5000, ""}, {"SP", "Season Pass", 2999, 1000, "All 9 nights"}}, Promoted: true, Hue: 330, Popularity: 80},
	{Code: "ET00520016", Title: "Football Fan Fest: Derby Screening", Category: "SPORTS", Genre: "Football", Languages: []string{"English"}, AgeLimit: "18yrs +", DurationMin: 180,
		Synopsis:  "Watch the big derby on a 40-ft screen with live commentary, face painting and the loudest crowd in town.",
		VenueName: "SkyDeck Arena", DayOffset: 6, Hour: 22, Tiers: []Tier{{"GA", "Fan Zone", 599, 600, ""}}, Hue: 140, Popularity: 54},
}

type Region struct {
	Code, Name, Slug, StateName, StateCode string
	Popular                                bool
	Lat, Lng                               float64
}

var Regions = []Region{
	{"MUMBAI", "Mumbai", "mumbai", "Maharashtra", "MH", true, 19.076, 72.8777},
	{"NCR", "Delhi-NCR", "national-capital-region-ncr", "Delhi", "DL", true, 28.6139, 77.209},
	{"BANG", "Bengaluru", "bengaluru", "Karnataka", "KA", true, 12.9716, 77.5946},
	{"HYD", "Hyderabad", "hyderabad", "Telangana", "TG", true, 17.385, 78.4867},
	{"CHD", "Chandigarh", "chandigarh", "Chandigarh", "CH", true, 30.7333, 76.7794},
	{"AHD", "Ahmedabad", "ahmedabad", "Gujarat", "GJ", true, 23.0225, 72.5714},
	{"PUNE", "Pune", "pune", "Maharashtra", "MH", true, 18.5204, 73.8567},
	{"CHEN", "Chennai", "chennai", "Tamil Nadu", "TN", true, 13.0827, 80.2707},
	{"KOLK", "Kolkata", "kolkata", "West Bengal", "WB", true, 22.5726, 88.3639},
	{"KOCH", "Kochi", "kochi", "Kerala", "KL", true, 9.9312, 76.2673},
	{"JAIP", "Jaipur", "jaipur", "Rajasthan", "RJ", false, 26.9124, 75.7873},
	{"LUCK", "Lucknow", "lucknow", "Uttar Pradesh", "UP", false, 26.8467, 80.9462},
	{"INDO", "Indore", "indore", "Madhya Pradesh", "MP", false, 22.7196, 75.8577},
	{"SURT", "Surat", "surat", "Gujarat", "GJ", false, 21.1702, 72.8311},
	{"NAGP", "Nagpur", "nagpur", "Maharashtra", "MH", false, 21.1458, 79.0882},
	{"COIM", "Coimbatore", "coimbatore", "Tamil Nadu", "TN", false, 11.0168, 76.9558},
	{"VIZA", "Visakhapatnam", "visakhapatnam", "Andhra Pradesh", "AP", false, 17.6868, 83.2185},
	{"BHUB", "Bhubaneswar", "bhubaneswar", "Odisha", "OD", false, 20.2961, 85.8245},
	{"GOA", "Goa", "goa", "Goa", "GA", false, 15.2993, 74.124},
	{"MYS", "Mysuru", "mysuru", "Karnataka", "KA", false, 12.2958, 76.6394},
}

// Chains are fictional multiplex brands. Tier scales ticket prices.
type Chain struct {
	Code, Name string
	Tier       float64
}

var Chains = []Chain{
	{"CSTR", "Cinestar", 1.15},
	{"GLXY", "Galaxy Multiplex", 1.0},
	{"RGLS", "Regal Screens", 0.85},
	{"MVDM", "Moviedome", 1.3},
	{"SLVR", "Silverline Cinemas", 0.75},
}

// Localities per region used to name venues.
var Localities = map[string][]string{
	"MUMBAI": {"Lower Parel", "Andheri (W)", "Ghatkopar", "Thane (W)", "Vashi", "Borivali (W)", "Kurla", "Malad"},
	"NCR":    {"Saket", "Connaught Place", "Vasant Kunj", "Gurugram Sector 29", "Noida Sector 18", "Rajouri Garden", "Dwarka"},
	"BANG":   {"Koramangala", "Whitefield", "Indiranagar", "Jayanagar", "Marathahalli", "Yelahanka", "Rajajinagar"},
	"HYD":    {"Banjara Hills", "Kukatpally", "Gachibowli", "Ameerpet", "Dilsukhnagar", "Kondapur"},
	"CHD":    {"Sector 17", "Elante", "Zirakpur", "Mohali Phase 8", "Panchkula"},
	"AHD":    {"SG Highway", "Prahlad Nagar", "Vastrapur", "Maninagar", "Motera"},
	"PUNE":   {"Koregaon Park", "Kothrud", "Viman Nagar", "Hinjewadi", "Camp", "Aundh"},
	"CHEN":   {"T. Nagar", "Velachery", "Anna Nagar", "OMR", "Vadapalani", "Guindy"},
	"KOLK":   {"Park Street", "Salt Lake", "New Town", "Howrah", "Gariahat", "Behala"},
	"KOCH":   {"Edappally", "MG Road", "Kakkanad", "Vyttila", "Aluva"},
	"JAIP":   {"Malviya Nagar", "C-Scheme", "Vaishali Nagar"},
	"LUCK":   {"Gomti Nagar", "Hazratganj", "Alambagh"},
	"INDO":   {"Vijay Nagar", "Palasia", "Rau"},
	"SURT":   {"Adajan", "Vesu", "Athwa"},
	"NAGP":   {"Sitabuldi", "Dharampeth", "Manish Nagar"},
	"COIM":   {"RS Puram", "Gandhipuram", "Peelamedu"},
	"VIZA":   {"Dwaraka Nagar", "MVP Colony", "Gajuwaka"},
	"BHUB":   {"Saheed Nagar", "Patia", "Jaydev Vihar"},
	"GOA":    {"Panaji", "Margao", "Mapusa"},
	"MYS":    {"Vijayanagar", "Kuvempunagar", "Saraswathipuram"},
}

// RegionLanguages biases which dubbed versions cinemas in a region screen.
var RegionLanguages = map[string][]string{
	"MUMBAI": {"Hindi", "English", "Marathi"}, "PUNE": {"Marathi", "Hindi", "English"}, "NAGP": {"Marathi", "Hindi"},
	"NCR": {"Hindi", "English", "Punjabi"}, "CHD": {"Hindi", "English", "Punjabi"}, "JAIP": {"Hindi", "English"},
	"LUCK": {"Hindi", "English"}, "INDO": {"Hindi", "English"},
	"BANG": {"Kannada", "English", "Hindi", "Tamil", "Telugu"}, "MYS": {"Kannada", "English"},
	"HYD": {"Telugu", "Hindi", "English"}, "VIZA": {"Telugu", "Hindi"},
	"CHEN": {"Tamil", "English", "Telugu"}, "COIM": {"Tamil", "English"},
	"KOLK": {"Bengali", "Hindi", "English"}, "BHUB": {"Odia", "Hindi", "English"},
	"KOCH": {"Malayalam", "Tamil", "English"}, "GOA": {"English", "Hindi", "Konkani"},
	"AHD": {"Gujarati", "Hindi", "English"}, "SURT": {"Gujarati", "Hindi"},
}
