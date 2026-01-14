// Package content provides narrative content for /dev/dungeon.
package content

import (
	"math/rand"
)

// ItemFlavorText contains extended descriptions for items.
var ItemFlavorText = map[string]ItemFlavor{
	// Consumables
	"sudo_potion": {
		Examine: `A shimmering vial of elevated privileges.

When consumed, this potion temporarily grants you root access,
making you immune to all damage. The label reads:

    "With great power comes great responsibility.
     Also, you'll probably break something."

Use wisely - or don't. You're root now.`,
		Use: []string{
			"You drink the sudo potion. Password accepted!",
			"Root privileges granted! You feel invincible!",
			"sudo echo 'I AM ROOT' >> /etc/motd",
			"You are now operating at ring 0. Nothing can touch you.",
			"Invincibility mode: ACTIVATED",
		},
	},
	"grep_scroll": {
		Examine: `A scroll covered in regular expressions.

When read aloud, this scroll reveals all items hidden on the
current floor. The incantation begins:

    "grep -r 'treasure' ./* 2>/dev/null"

Some say if you read it backwards, you can find things that
don't want to be found.`,
		Use: []string{
			"You read the grep scroll. Items revealed!",
			"grep -r '.*' ./ | Items light up on your map!",
			"The scroll burns up after revealing all secrets.",
			"Pattern matched! All items now visible!",
			"You grep the floor. Nothing is hidden from you.",
		},
	},
	"malloc": {
		Examine: `A crystallized memory allocation call.

This fragment of pure memory can be consumed to restore
your RAM. Handle with care - memory management is no joke.

    void* health = malloc(30 * sizeof(HP));
    if (health == NULL) { /* you're already dead */ }

Side effects may include: buffer overflows, use-after-free,
and that one bug you can never reproduce.`,
		Use: []string{
			"malloc(30) successful! RAM restored!",
			"Memory allocated. You feel more... spacious.",
			"Your heap grows. +30 RAM!",
			"You consume the malloc. Don't forget to free() later!",
			"Memory successfully allocated to your process.",
		},
	},
	"fd_restore": {
		Examine: `A small utility for closing file descriptors.

Your process can only have so many files open at once.
This handy tool closes the ones you forgot about.

    for (int fd = 3; fd < MAX_FD; fd++) close(fd);

Remember: every open file is a potential resource leak.
Your process thanks you for being responsible.`,
		Use: []string{
			"close() successful! FDs restored!",
			"File descriptors closed. Resources freed!",
			"You close your unused FDs. +4 FD capacity!",
			"ulimit -n increased! More room for abilities!",
			"FD leak plugged. You can breathe again.",
		},
	},
	// Weapons
	"kill_9": {
		Examine: `The legendary kill -9 command, forged into weapon form.

This is the nuclear option. SIGKILL cannot be caught, cannot
be blocked, cannot be ignored. When you absolutely, positively
need to kill every process in the room.

    kill -9 -- because sometimes SIGTERM just won't cut it.

WARNING: May cause orphaned processes. Use responsibly.
(You won't.)`,
		Equip: []string{
			"You equip kill -9. Nothing is safe now.",
			"The weight of SIGKILL settles in your hand.",
			"Signal 9 ready. Time to terminate some processes.",
			"kill -9 equipped. Judge, jury, and executioner.",
			"With this weapon, you ARE the OOM killer.",
		},
	},
	"basic_script": {
		Examine: `A simple bash script, crudely weaponized.

#!/bin/bash
# basic_attack.sh
echo "Attacking target..."
damage=3
echo "Deal $damage damage"

It's not elegant, but it gets the job done. Mostly.
Sometimes. When the permissions are right.`,
		Equip: []string{
			"You equip the bash script. chmod +x and ready!",
			"Basic but reliable. The script is armed.",
			"#!/bin/bash equipped. Time to automate violence.",
			"You wield the script. It's POSIX-compliant!",
			"Script loaded. May the shebang be with you.",
		},
	},
	// Armor
	"chmod_x": {
		Examine: `Execution permissions crystallized into armor.

When worn, this artifact increases your RAM and lowers your
UID, granting you elevated privileges. The inscription reads:

    chmod +x /usr/bin/you

With execute permissions comes the ability to run yourself.
Deep, when you think about it. Also, more health.`,
		Equip: []string{
			"chmod +x applied. You are now executable.",
			"Permissions elevated. UID decreased!",
			"Execute permissions granted. +20 RAM!",
			"You don the chmod armor. -rwxr-xr-x achieved.",
			"The armor grants you the right to execute.",
		},
	},
	"basic_shell": {
		Examine: `A protective shell. Literally.

/bin/sh has protected processes since the dawn of Unix.
Now it can protect you too. It's not bash, but it's portable.

    $ # I am a shell. Fear me.

Some say zsh is better. Those people are wrong.
(Don't @ me.)`,
		Equip: []string{
			"You enter the shell. Protection enabled.",
			"/bin/sh wraps around you. Cozy.",
			"Shell equipped. You feel... sheltered.",
			"The basic shell provides basic protection.",
			"sh> Ready to take some hits.",
		},
	},
	// Loot drops
	"memory_fragment": {
		Examine: `A fragment of discarded memory.

Once part of a larger allocation, this shard of RAM floats
through the system looking for a process to call home.
You can absorb it to restore some of your health.

    // TODO: figure out where this memory came from
    // TODO: find the memory leak
    // TODO: fix the memory leak (lol)`,
		Use: []string{
			"You absorb the memory fragment. +10 RAM!",
			"The RAM merges with your process. Absorbed!",
			"Memory reclaimed! You feel more substantial.",
			"Fragment integrated. Your heap grows.",
			"Lost memory returned to a good home.",
		},
	},
	"service_token": {
		Examine: `A systemd service token from a fallen daemon.

This token once allowed a daemon to interact with the
init system. Now it can fuel your process with +20 RAM.

    [Unit]
    Description=Fallen Daemon's Token
    After=you-pick-it-up.target

The systemd controversy continues even in death.`,
		Use: []string{
			"Token consumed! +20 RAM allocated!",
			"systemctl start restoration. +20 RAM!",
			"The daemon's essence restores you.",
			"Service token activated. Health restored!",
			"You absorb the daemon's power. +20 RAM!",
		},
	},
	"cpu_cycle": {
		Examine: `A crystallized CPU cycle stolen from a fork bomb.

This pure computational cycle can temporarily boost your
attack power. One cycle might not seem like much, but at
the right frequency, it's devastating.

    while(true) { attack++; }

Handle with care. Do not overclock.`,
		Use: []string{
			"CPU cycle consumed! +5 CPU for this combat!",
			"Your clock speed increases! Temporary power!",
			"Cycle integrated. Attacking faster!",
			"CPU boosted! Time to deal some damage!",
			"Overclocking engaged! +5 attack power!",
		},
	},
	"core_dump": {
		Examine: `The compressed remains of a crashed process.

Inside this core dump is the last memory state of something
that died badly. You can weaponize this pain, dealing 40
damage to an enemy.

    Program terminated with signal SIGSEGV
    #0  0xdeadbeef in tragedy ()

Someone's bug is your weapon.`,
		Use: []string{
			"You hurl the core dump! 40 damage!",
			"Core dump deployed! Memories weaponized!",
			"The crash data strikes true! 40 damage!",
			"Enemy receives a face full of stack trace!",
			"Segfault delivered! The enemy takes 40 damage!",
		},
	},
	"root_shard": {
		Examine: `A fragment of root privilege.

This rare shard permanently lowers your UID, bringing you
closer to root access. The closer to 0, the more powerful
you become.

    # echo "I am closer to root" >> /etc/passwd

Lower UID = higher privilege. Welcome to Unix.`,
		Use: []string{
			"Root shard absorbed! UID permanently decreased!",
			"You feel more... privileged. -100 UID!",
			"The shard merges with your process. Power flows!",
			"UID decreased! You're closer to root!",
			"Privilege escalation successful! Permanent boost!",
		},
	},
}

// ItemFlavor contains flavor text for a single item.
type ItemFlavor struct {
	Examine string   // Text when examining the item
	Use     []string // Random texts when using the item
	Equip   []string // Random texts when equipping the item
}

// GetItemExamine returns the examine text for an item.
func GetItemExamine(itemID string) string {
	if flavor, ok := ItemFlavorText[itemID]; ok {
		return flavor.Examine
	}
	return "A mysterious item. Its purpose is unclear."
}

// GetItemUseMessage returns a random use message for an item.
func GetItemUseMessage(itemID string) string {
	if flavor, ok := ItemFlavorText[itemID]; ok && len(flavor.Use) > 0 {
		return flavor.Use[rand.Intn(len(flavor.Use))]
	}
	return "You use the item. Something happens!"
}

// GetItemEquipMessage returns a random equip message for an item.
func GetItemEquipMessage(itemID string) string {
	if flavor, ok := ItemFlavorText[itemID]; ok && len(flavor.Equip) > 0 {
		return flavor.Equip[rand.Intn(len(flavor.Equip))]
	}
	return "You equip the item. It feels right."
}

// RarityNames maps rarity levels to display names.
var RarityNames = map[int]string{
	0: "Common",
	1: "Uncommon",
	2: "Rare",
	3: "Epic",
	4: "Legendary",
}

// RarityColors maps rarity levels to ANSI color codes.
var RarityColors = map[int]string{
	0: "\033[37m",  // White
	1: "\033[32m",  // Green
	2: "\033[34m",  // Blue
	3: "\033[35m",  // Purple/Magenta
	4: "\033[33m",  // Gold/Yellow
}

// ItemPickupMessages when picking up items.
var ItemPickupMessages = []string{
	"Item acquired! Added to inventory.",
	"You pick up the item. Inventory updated.",
	"Got it! The item is now yours.",
	"Item claimed! Check your inventory.",
	"Loot secured! Item added to your stash.",
	"You grab the item before it despawns.",
}

// GetPickupMessage returns a random pickup message.
func GetPickupMessage() string {
	return ItemPickupMessages[rand.Intn(len(ItemPickupMessages))]
}

// InventoryFullMessages when inventory is full.
var InventoryFullMessages = []string{
	"Inventory full! Cannot carry more items.",
	"Your /proc/self/fd is maxed out! Drop something!",
	"No room! Inventory at capacity!",
	"Cannot carry more! Your hands are full!",
	"Max items reached! Make some room first!",
}

// GetInventoryFullMessage returns a random inventory full message.
func GetInventoryFullMessage() string {
	return InventoryFullMessages[rand.Intn(len(InventoryFullMessages))]
}

// ItemDropMessages when dropping items.
var ItemDropMessages = []string{
	"Item dropped. Someone else's problem now.",
	"You leave the item behind.",
	"Item discarded. It clatters to the ground.",
	"Freed from your inventory. The item awaits.",
	"You drop the item. Goodbye, old friend.",
}

// GetDropMessage returns a random drop message.
func GetDropMessage() string {
	return ItemDropMessages[rand.Intn(len(ItemDropMessages))]
}
