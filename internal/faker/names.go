package faker

import (
	"strings"
)

var firstNames = []string{
	"Alexander", "Adrian", "Andrei", "Bogdan", "Catalin", "Cristian", "Dan", "Daniel",
	"Elena", "Gabriel", "Ioana", "Ion", "Maria", "Mihai", "Radu", "Stefan",
	"James", "Mary", "John", "Patricia", "Robert", "Jennifer", "Michael", "Linda",
	"William", "Elizabeth", "David", "Barbara", "Richard", "Susan", "Joseph", "Jessica",
	"Thomas", "Sarah", "Charles", "Karen", "Christopher", "Nancy", "Daniel", "Lisa",
	"Matthew", "Betty", "Anthony", "Margaret", "Mark", "Sandra", "Donald", "Ashley",
	"Steven", "Kimberly", "Paul", "Emily", "Andrew", "Donna", "Joshua", "Michelle",
	"Kenneth", "Dorothy", "Kevin", "Carol", "Brian", "Amanda", "George", "Melissa",
	"Edward", "Deborah", "Ronald", "Stephanie", "Timothy", "Rebecca", "Jason", "Sharon",
	"Jeffrey", "Laura", "Ryan", "Cynthia", "Jacob", "Kathleen", "Gary", "Amy",
	"Nicholas", "Angela", "Eric", "Shirley", "Jonathan", "Emma", "Stephen", "Brenda",
	"Larry", "Pamela", "Justin", "Nicole", "Scott", "Anna", "Brandon", "Samantha",
	"Benjamin", "Katherine", "Samuel", "Christine", "Gregory", "Debra", "Frank", "Rachel",
	"Lucas", "Emma", "Liam", "Olivia", "Noah", "Sophia", "Oliver", "Ava",
	"Elijah", "Isabella", "Mateo", "Mia", "Arthur", "Chloe", "Leo", "Camille",
}

var lastNames = []string{
	"Popescu", "Ionescu", "Radu", "Dumitrescu", "Stoica", "Gheorghe", "Stan", "Matei",
	"Smith", "Johnson", "Williams", "Brown", "Jones", "Garcia", "Miller", "Davis",
	"Rodriguez", "Martinez", "Hernandez", "Lopez", "Gonzalez", "Wilson", "Anderson", "Thomas",
	"Taylor", "Moore", "Jackson", "Martin", "Lee", "Perez", "Thompson", "White",
	"Harris", "Sanchez", "Clark", "Ramirez", "Lewis", "Robinson", "Walker", "Young",
	"Allen", "King", "Wright", "Scott", "Torres", "Nguyen", "Hill", "Flores",
	"Green", "Adams", "Nelson", "Baker", "Hall", "Rivera", "Campbell", "Mitchell",
	"Carter", "Roberts", "Gomez", "Phillips", "Evans", "Turner", "Diaz", "Parker",
	"Cruz", "Edwards", "Collins", "Reyes", "Stewart", "Morris", "Morales", "Murphy",
	"Cook", "Rogers", "Gutierrez", "Ortiz", "Morgan", "Cooper", "Peterson", "Bailey",
	"Reed", "Kelly", "Howard", "Ramos", "Kim", "Cox", "Ward", "Richardson",
	"Watson", "Brooks", "Chavez", "Wood", "James", "Bennett", "Gray", "Mendoza",
	"Ruiz", "Hughes", "Price", "Alvarez", "Castillo", "Sanders", "Patel", "Myers",
	"Long", "Ross", "Foster", "Jimenez", "Powell", "Jenkins", "Perry", "Russell",
	"Sullivan", "Bell", "Coleman", "Butler", "Henderson", "Barnes", "Gonzales", "Fisher",
}

// GenerateFirstName returns a deterministic first name.
func GenerateFirstName(rng *DeterministicPRNG) string {
	return rng.PickString(firstNames)
}

// GenerateLastName returns a deterministic last name.
func GenerateLastName(rng *DeterministicPRNG) string {
	return rng.PickString(lastNames)
}

// GenerateFullName returns a combined deterministic first and last name.
func GenerateFullName(rng *DeterministicPRNG) string {
	return GenerateFirstName(rng) + " " + GenerateLastName(rng)
}

// GenerateUsername generates a deterministic handle like "alex.smith42".
func GenerateUsername(rng *DeterministicPRNG) string {
	fn := strings.ToLower(GenerateFirstName(rng))
	ln := strings.ToLower(GenerateLastName(rng))
	suffix := rng.IntBetween(10, 99)
	return fn + "." + ln + string(rune('0'+suffix/10)) + string(rune('0'+suffix%10))
}
