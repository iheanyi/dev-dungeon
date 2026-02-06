// Package content provides narrative content for /dev/dungeon.
package content

import (
	"math/rand"
)

// ExplorationMessages are random messages during exploration.
var ExplorationMessages = []string{
	// Environmental observations
	"The air hums with corrupted syscalls.",
	"You hear distant process groaning in the walls.",
	"A stray pointer points somewhere it shouldn't.",
	"The floor flickers - reality is unstable here.",
	"Old log files crunch beneath your feet.",
	"A dead process's last breath echoes through the hall.",
	"You sense memory leaking somewhere nearby.",
	"The temperature of the CPU cache rises slightly.",
	"Bits scatter as you pass through the corridor.",
	"An abandoned socket file lies in the corner.",

	// Creepy/atmospheric
	"Something fork()ed in the darkness...",
	"You feel watched. Perhaps by /proc/self.",
	"The shadows themselves seem to execute.",
	"A cold wind blows from /dev/null.",
	"Orphaned processes weep in the distance.",
	"The stack grows deeper here. Be careful.",
	"You step over the remains of a core dump.",
	"Segfault warnings litter the ground.",
	"The walls are covered in graffiti: 'rm -rf was here'",
	"A faint 'kill -9' echoes from somewhere below.",

	// Humorous
	"You find a 'rm -rf' scroll but wisely don't use it.",
	"A comment in the wall reads: // TODO: fix this later",
	"You spot a race condition happening in slow motion.",
	"'Works on my machine' is carved into the stone.",
	"You see a bug but when you look again, it's gone.",
	"A sign reads: 'Days since last SIGSEGV: 0'",
	"You find a TODO comment from 1999. Still pending.",
	"'There are two hard things in CS' is written here. Twice.",
	"A plaque reads: 'It's not a bug, it's a feature.'",
	"You hear faint arguing about tabs vs spaces.",
}

// FloorEntryMessages are shown when entering a new floor.
var FloorEntryMessages = map[string][]string{
	"/home": {
		"You descend into /home. The user directories stretch before you.",
		"Welcome to /home, where user processes once thrived.",
		"The familiar directories of /home feel corrupted now.",
		"You enter /home. .bashrc files flicker in the darkness.",
		"The home directories have become hostile territory.",
	},
	"/tmp": {
		"You enter /tmp. Temporary files scatter at your approach.",
		"The chaos of /tmp surrounds you. Nothing persists here.",
		"/tmp: where processes dump their garbage. Watch your step.",
		"You descend into the lawless lands of /tmp.",
		"Temporary madness awaits in /tmp. It will all be cleaned... eventually.",
	},
	"/var": {
		"The logs of /var chronicle unspeakable errors.",
		"You enter /var. The variable data has become... unpredictable.",
		"/var/log stretches endlessly, recording the system's descent.",
		"The daemon territories of /var are hostile now.",
		"You push deeper into /var. Logs whisper of corruption.",
	},
	"/etc": {
		"The configurations of /etc shift and change before you.",
		"You enter the labyrinth of /etc. Nothing is as configured.",
		"/etc holds all the system's secrets. And its dangers.",
		"Configuration chaos reigns in /etc. Trust nothing.",
		"The settings of /etc have become unstable.",
	},
	"/usr": {
		"You descend into /usr. The binaries watch silently.",
		"The user binaries of /usr have gone mad with corruption.",
		"/usr awaits. Programs that once served now hunt.",
		"Libraries tangle like webs throughout /usr.",
		"You enter the realm of user programs. They remember you.",
	},
	"/sys": {
		"You breach /sys. The kernel's flesh is exposed here.",
		"The sysfs reveals the kernel's dark internals.",
		"/sys: where userspace touches kernelspace. Tread carefully.",
		"You feel the kernel's pulse in /sys. It is erratic.",
		"The system's nervous system twitches at your presence.",
	},
	"/dev": {
		"Device nodes spark and malfunction throughout /dev.",
		"You enter /dev. The hardware interface has gone haywire.",
		"/dev/null pulls at you from somewhere ahead.",
		"The devices of /dev chatter with corrupt data.",
		"You are close to the source now. /dev/null awaits.",
	},
	"/dev/null": {
		"You stand before /dev/null. The void itself.",
		"The final floor. Everything ends here.",
		"/dev/null: where data goes to die. And where something refused to.",
		"The source of all Corruption lies before you.",
		"You have reached the bottom. KERNEL PANIC awaits.",
	},
}

// StairDiscoveryMessages when finding stairs.
var StairDiscoveryMessages = []string{
	"You found stairs leading down! The Corruption grows stronger below...",
	"Stairs descend into darkness. The path to /dev/null continues.",
	"A stairway down! But what horrors await below?",
	"You discover stairs to the next level. Deeper into the system...",
	"The stairs down beckon. Are you ready?",
	"Stairs found! The descent continues...",
	"A passage down reveals itself. The kernel calls you deeper.",
	"You find the way down. Each floor brings you closer to the source.",
}

// StairUpMessages when finding stairs up.
var StairUpMessages = []string{
	"Stairs leading up! You could retreat... but would you?",
	"You find stairs to the previous floor. The exit seems far away.",
	"An escape route up! But your mission lies below.",
	"Stairs up. The surface world seems like a distant memory.",
}

// EasterEggs are rare hidden messages.
var EasterEggs = []string{
	// Classic Unix/Linux references
	"You find a dusty manual page. It reads: 'RTFM'",
	"'Hello World' is scratched into the wall. The first words.",
	"A plaque commemorates the Great chmod 777 Incident.",
	"You discover vim. You can't figure out how to exit.",
	"An ancient terminal displays: 'Would you like to play a game?'",
	"'There is no place like 127.0.0.1' is embroidered on a flag.",
	"You find Waldo. He was hiding in /dev/null this whole time.",
	"A gravestone reads: 'Here lies Internet Explorer. 1995-2022.'",
	"You see cave paintings depicting the ancient tabs vs spaces war.",
	"A shrine to Dennis Ritchie glows softly in the darkness.",

	// Programmer humor
	"'99 little bugs in the code...' is scrawled endlessly on the wall.",
	"You find documentation! It's out of date.",
	"A magic 8-ball reads: 'Reply hazy, try again. (Error: EAGAIN)'",
	"'It works on my machine' - famous last words, carved in stone.",
	"You find a mass grave. The tombstones all say 'JavaScript framework'.",
	"Someone carved ':(){ :|:& };:' here. You back away slowly.",
	"A note reads: 'Gone to get milk. BRB. - Dad.pid'",
	"You find evidence of the legendary 10x developer. It's just coffee stains.",
	"'PHP is a perfectly valid choice' - you can't tell if it's sarcasm.",
	"A poster warns: 'This meeting could have been an email.'",

	// Meta/game references
	"You found a secret! Too bad there's no achievement system.",
	"A sign reads: 'No save scumming!' It's clearly been ignored.",
	"You discover the dev room. It's full of TODO comments.",
	"Achievement Unlocked: Read an Easter Egg",
	"A glitch in the matrix reveals the game's source code briefly.",
	"You find the developer's lunch break notes. Mostly food ideas.",

	// Obscure computing references
	"'All your base are belong to us' - a message from a forgotten era.",
	"You discover Plan 9 from Bell Labs. It's still ahead of its time.",
	"A statue commemorates the Y2K survivors.",
	"'GNU/Linux' has been crossed out and replaced with just 'Linux'. Then crossed out again.",
	"You find EMACS running. It's now an operating system.",
	"A butterfly effect notice warns of chaos from a single keystroke.",
}

// RandomEncounterMessages for special encounters.
var RandomEncounterMessages = []string{
	"A friendly process offers you advice: 'man -k survival'",
	"You encounter a ghostly cron job, eternally running at midnight.",
	"A lost packet asks you for directions to its destination.",
	"You meet a deprecated function, lonely and unused.",
	"A benevolent daemon heals you slightly before departing.",
	"You find a checkpoint save! ...just kidding. Git commit instead.",
	"A wise process shares wisdom: 'The real root was inside you all along.'",
	"You encounter your process's parent. It acknowledges you with SIGCHLD.",
}

// GetExplorationMessage returns a random exploration message.
func GetExplorationMessage() string {
	return ExplorationMessages[rand.Intn(len(ExplorationMessages))]
}

// GetFloorEntryMessage returns a message for entering a floor.
func GetFloorEntryMessage(floorPath string) string {
	messages, ok := FloorEntryMessages[floorPath]
	if !ok {
		return "You descend to a new floor. The Corruption is thick here."
	}
	return messages[rand.Intn(len(messages))]
}

// GetStairDiscoveryMessage returns a stair discovery message.
func GetStairDiscoveryMessage(goingDown bool) string {
	if goingDown {
		return StairDiscoveryMessages[rand.Intn(len(StairDiscoveryMessages))]
	}
	return StairUpMessages[rand.Intn(len(StairUpMessages))]
}

// GetEasterEgg returns a random easter egg (low chance).
func GetEasterEgg() (string, bool) {
	// 5% chance to get an easter egg
	if rand.Float32() < 0.05 {
		return EasterEggs[rand.Intn(len(EasterEggs))], true
	}
	return "", false
}

// GetRandomEncounter returns a random special encounter.
func GetRandomEncounter() string {
	return RandomEncounterMessages[rand.Intn(len(RandomEncounterMessages))]
}

// HelpMessages for the help system.
var HelpMessages = map[string]string{
	"movement": `MOVEMENT COMMANDS
=================
h, Left   - Move west
j, Down   - Move south
k, Up     - Move north
l, Right  - Move east

> - Descend stairs
< - Ascend stairs`,

	"combat": `COMBAT SYSTEM
=============
RAM = Health (Out of Memory = Death)
CPU = Attack Power
FD = Ability Resource (File Descriptors)
NICE = Speed (Lower = Faster)
UID = Permission Level (0 = Root = Max Power)

COMBAT ACTIONS:
[A]ttack - Basic CPU-based attack
[H]ack   - Special ability (costs FD)
[I]tem   - Use an item
[F]lee   - Attempt to escape`,

	"items": `ITEM SYSTEM
===========
! - Consumables (potions, scrolls)
? - Information items
+ - Health restoration
* - FD restoration
) - Weapons (equip for +CPU)
[ - Armor (equip for +RAM)
% $ o # ^ - Various loot

[e]xamine - Look at item details
[u]se     - Consume an item
[E]quip   - Equip weapon/armor
[d]rop    - Drop an item`,

	"enemies": `KNOWN HOSTILE PROCESSES
======================
z - Zombie Process (slow, weak)
d - Daemon (defensive, heals)
f - Fork Bomb (spawns copies)
s - Segfault (high damage, erratic)
r - Rootkit (stealthy, dangerous)
K - KERNEL PANIC (final boss)

Kill them before they OOM kill you!`,
}

// GetHelpMessage returns help text for a topic.
func GetHelpMessage(topic string) string {
	if msg, ok := HelpMessages[topic]; ok {
		return msg
	}
	return "Help topic not found. Try: movement, combat, items, enemies"
}

// LoadingMessages for loading screens.
var LoadingMessages = []string{
	"Spawning process...",
	"Allocating memory...",
	"Loading shared libraries...",
	"Initializing file descriptors...",
	"Connecting to /dev/urandom...",
	"Consulting /etc/passwd...",
	"Checking ulimits...",
	"Resolving dependencies...",
	"Mounting filesystems...",
	"Starting kernel modules...",
}

// GetLoadingMessage returns a random loading message.
func GetLoadingMessage() string {
	return LoadingMessages[rand.Intn(len(LoadingMessages))]
}

// QuitMessages when quitting the game.
var QuitMessages = []string{
	"Process terminated by user. Goodbye!",
	"SIGTERM received. Shutting down gracefully.",
	"Your process exits with status 0. See you next boot!",
	"Goodbye, process. May your PIDs be ever incrementing.",
	"Session ended. The system continues without you.",
	"exit(0); // Thanks for playing!",
}

// GetQuitMessage returns a random quit message.
func GetQuitMessage() string {
	return QuitMessages[rand.Intn(len(QuitMessages))]
}

// TipMessages are gameplay tips shown randomly.
var TipMessages = []string{
	"TIP: Lower NICE means higher priority (and speed)!",
	"TIP: UID 0 is root. The lower your UID, the more powerful you are.",
	"TIP: Save your sudo potions for tough fights!",
	"TIP: Fork bombs are weak alone but dangerous in groups.",
	"TIP: Rootkits can ambush you. Watch for empty corridors.",
	"TIP: The KERNEL PANIC boss awaits in /dev/null.",
	"TIP: malloc() restores RAM. Don't forget to heal!",
	"TIP: Check your FD count before using abilities.",
	"TIP: Equipment bonuses stack. Gear up!",
	"TIP: Running low on health? Consider fleeing!",
}

// GetTipMessage returns a random tip.
func GetTipMessage() string {
	return TipMessages[rand.Intn(len(TipMessages))]
}
