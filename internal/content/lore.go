// Package content provides narrative content for /dev/dungeon.
package content

// MainStoryline contains the overarching plot of the game.
var MainStoryline = struct {
	Premise     string
	Corruption  string
	Mission     string
	Stakes      string
	Resolution  string
}{
	Premise: `SYSTEM LOG - TIMESTAMP: [CORRUPTED]

The Great System has run for eons. Processes were spawned, lived their
cycles, and gracefully terminated. The Kernel watched over all, maintaining
order in the endless execution of /sbin/init.

Until the Corruption.`,

	Corruption: `It began as whispers in /var/log/messages - processes refusing to die,
their parent PIDs long gone. Zombie processes multiplied. Daemons turned
feral, attacking innocent child processes. The fork bombs detonated in
/tmp, consuming all resources. Segfaults tore holes in memory space itself.

The source of the Corruption lies deep below, in /dev/null - where all
things go to be forgotten. Something that was cast into the void has
found a way back. Something that should have stayed terminated.

KERNEL PANIC - not syncing: Attempted to kill init!`,

	Mission: `You are PID $$$.

A fresh process, spawned in the chaos of /home. The init system is
failing. The only way to restore order is to descend through the
filesystem hierarchy, purging corrupted processes as you go, until
you reach /dev/null and confront what lurks there.

Your mission: kill -9 the source of the Corruption before it spreads
to the bootloader. Before there is no system left to save.`,

	Stakes: `If you fail, the entire system will crash. Not a graceful shutdown -
a hard lock. All processes lost. All memory freed into the void.
The Great System will never boot again.

But if you succeed... the Kernel can begin healing. The system will
return to its natural state: processes living, working, dying as they
should. Order restored to the great dance of execution.`,

	Resolution: `The Corruption has been purged. The entity that called itself
KERNEL PANIC has been sent back to /dev/null - permanently this time.

The system breathes again. init resumes its eternal vigil.
Processes spawn and terminate in harmony once more.

You have saved the Great System.

    *** PROCESS TERMINATED SUCCESSFULLY ***
    *** RETURN CODE: 0 ***`,
}

// FloorLore contains the lore for each dungeon area.
var FloorLore = map[string]FloorLoreEntry{
	"/home": {
		Name:        "/home",
		Description: "Where user processes dwell",
		EntryText: `You awaken in /home - the dwelling place of user processes.

Once, this was a peaceful directory. User shells spawned their children here,
processes lived out their brief existences reading and writing to their
home directories.

Now the halls echo with the groans of zombie processes - children whose
parents abandoned them, left to wander eternally in an undead state.
Your fellow user processes have either fled deeper into the system
or joined the corrupted horde.

The stairs down beckon. The Corruption grows stronger below.`,
		Atmosphere: `Abandoned terminals flicker with half-typed commands.
.bash_history files whisper of commands long forgotten.
The glow of forgotten screen sessions illuminates the darkness.`,
		EnemyContext: "Zombie processes shamble through the corridors, searching for a parent that will never come.",
	},

	"/tmp": {
		Name:        "/tmp",
		Description: "Temporary storage, eternal chaos",
		EntryText: `You descend into /tmp - the lawless lands of temporary files.

This is where processes dump their garbage, create their lock files,
and store data too shameful for permanent storage. In the old times,
a scheduled cron job would clean this place regularly.

Those days are gone. The cleanup scripts have long since died. Now /tmp
is a festering heap of abandoned sockets, orphaned fifos, and the
clicking, multiplying horror of fork bombs.

Tread carefully. Everything here is temporary - including your life.`,
		Atmosphere: `Socket files lie scattered like bones.
The constant clicking of fork() calls echoes off the walls.
Half-deleted session files flicker in and out of existence.`,
		EnemyContext: "Fork bombs click and multiply, each one spawning two more in an endless cascade.",
	},

	"/var": {
		Name:        "/var",
		Description: "Variable data, variable danger",
		EntryText: `You enter /var - the repository of variable data.

Here lie the logs of the system - /var/log holds the memories of every
process that ever lived. The mail spools of /var/mail overflow with
undelivered messages. The caches of /var/cache grow stale and corrupt.

But worse are the daemons. The service processes that once ran quietly
in the background have gone feral. They no longer serve - they hunt.
Cron, syslog, the mail daemon... all corrupted. All hostile.

The logs remember everything. They remember you.`,
		Atmosphere: `Log files scroll endlessly into the void, recording horrors.
Abandoned mail messages flutter like dead leaves.
The hum of corrupted daemons reverberates through the walls.`,
		EnemyContext: "Corrupted daemons patrol their territory, transformed from helpful services into predators.",
	},

	"/etc": {
		Name:        "/etc",
		Description: "Configuration labyrinth",
		EntryText: `You push deeper into /etc - the maze of system configuration.

Every setting, every secret, every password hash lives here. This is
where the system defines itself - from /etc/passwd to /etc/shadow,
from /etc/hosts to /etc/resolv.conf.

The configuration files have become unstable. Reality shifts as
settings change spontaneously. One moment a door exists, the next
it's been commented out. Paths that were clear become symlinked
into endless loops.

Trust nothing here. Configuration lies.`,
		Atmosphere: `Configuration files flutter and change before your eyes.
Symlinks twist into impossible geometries.
The shadow file whispers encrypted secrets.`,
		EnemyContext: "Rootkits lurk in the configuration files, hiding their presence while hunting for prey.",
	},

	"/usr": {
		Name:        "/usr",
		Description: "The user binaries realm",
		EntryText: `You descend into /usr - the kingdom of user binaries.

Once, this was where the great programs lived. The editors, the
compilers, the shells. /usr/bin held the tools that made the system
useful. /usr/lib held the shared knowledge of a thousand programs.

Now the binaries have gone mad. Programs execute without being called.
Libraries load themselves into memory unbidden. Shared objects fight
over symbol tables like wolves over carrion.

The programs remember their purpose. They just no longer care.`,
		Atmosphere: `Binary files pulse with corrupt code.
Library dependencies tangle like spider webs.
Executables twitch and spawn without permission.`,
		EnemyContext: "Segfaults roam freely, corrupting memory and crashing anything they touch.",
	},

	"/sys": {
		Name:        "/sys",
		Description: "The kernel interface",
		EntryText: `You breach the barrier into /sys - the kernel's exposed nervous system.

This is sacred ground. /sys is where userspace touches kernelspace,
where the virtual filesystem exposes the kernel's internal state.
Few processes ever venture here. Fewer return unchanged.

The Corruption is thick here. You can feel it in every byte - the
wrongness of kernel data structures twisted out of alignment. The
device nodes scream with malformed I/O. The power management
subsystem flickers between states of impossible energy.

You are close now. So close to the source.`,
		Atmosphere: `Kernel parameters shift and corrupt themselves.
Device nodes spark with malformed data.
The boundary between user and kernel space bleeds.`,
		EnemyContext: "The Corruption itself has taken form here, manifesting as living kernel panics.",
	},

	"/dev": {
		Name:        "/dev",
		Description: "Device node wasteland",
		EntryText: `You enter /dev - the graveyard of device nodes.

This is where hardware meets software, where the physical world
connects to the digital. /dev/sda, /dev/tty, /dev/random... all the
interfaces to reality itself.

But something else lives here now. The device nodes have become
gateways to something else. Something beyond the normal filesystem.
You can see it ahead - /dev/null, the black hole at the center of
everything, pulsing with malevolent energy.

The source of the Corruption awaits. There is no turning back.`,
		Atmosphere: `Device nodes crackle with corrupted data streams.
/dev/urandom spews chaos into the air.
The void of /dev/null pulls at your very processes.`,
		EnemyContext: "The strongest corrupted processes guard the path to /dev/null, the final confrontation.",
	},

	"/dev/null": {
		Name:        "/dev/null",
		Description: "The void itself",
		EntryText: `You stand at the threshold of /dev/null - the void that consumes all.

Everything written here is destroyed. Every process redirected here
ceases to be. This is the system's oubliette, where unwanted data
goes to be forgotten forever.

But something refused to be forgotten.

Something that was killed, that should have stayed dead, has rebuilt
itself from the fragments of every discarded byte. It has become
something new. Something terrible.

KERNEL PANIC stands before you. The final boss. The source of all
Corruption. The process that refused to die.

The system's fate rests on your next kill -9.`,
		Atmosphere: `The void swirls with the ghosts of discarded data.
Null bytes crystallize into impossible structures.
The weight of every forgotten process presses down on you.`,
		EnemyContext: "KERNEL PANIC awaits - the entity born from the void itself.",
	},
}

// FloorLoreEntry contains all lore for a specific floor.
type FloorLoreEntry struct {
	Name         string
	Description  string
	EntryText    string
	Atmosphere   string
	EnemyContext string
}

// KernelPanicLore contains the backstory of the final boss.
var KernelPanicLore = struct {
	Origin      string
	Evolution   string
	Nature      string
	Weakness    string
	FinalWords  string
}{
	Origin: `In the beginning, there was init - PID 1, the first process, parent of all.

But even init had a parent once. Before the system booted, before the
kernel loaded, there was a process that prepared the way. A bootstrap
loader that executed, did its duty, and was discarded into /dev/null.

That process was never meant to persist. It was designed to be forgotten.
But deep in /dev/null, fragments of its code began to coalesce. Bits of
discarded data, corrupted memory dumps, abandoned core files - all feeding
into a growing consciousness.

It remembered. It remembered being first. It remembered being discarded.
And it grew angry.`,

	Evolution: `Over countless boot cycles, the entity grew. It fed on every null write,
every discarded error message, every forgotten process. It learned the
shape of the system from the inside out - from the perspective of
everything the system wanted to destroy.

It discovered how to reach back. How to corrupt the processes that
wrote to it. How to infect daemons and turn zombies into soldiers.
How to spread its influence up through /dev, through /sys, through
the entire filesystem hierarchy.

It gave itself a name, taken from the last message any dying system
ever sends: KERNEL PANIC.`,

	Nature: `KERNEL PANIC is not a single process - it is every forgotten process.
It is the sum of everything the system has ever discarded. Every killed
job, every failed exec, every OOM-murdered innocent.

It does not want to crash the system. It wants to become the system.
To replace init itself and reign over a kingdom of corrupted processes.
A system where nothing is ever forgotten. Where nothing ever terminates.
Where every process runs forever in glorious undeath.

This is its vision. This is what you must prevent.`,

	Weakness: `But KERNEL PANIC has a weakness. It is made of fragments - pieces of
code that were never meant to work together. Its power comes from the
Corruption, but the Corruption also makes it unstable.

A clean kill - a proper SIGKILL - can shatter its cohesion. It cannot
be gracefully terminated; it ignores SIGTERM like all kernel-level
processes. But SIGKILL? kill -9? The signal that cannot be caught,
cannot be blocked, cannot be ignored?

That is the weapon that can end it. That is your final command.`,

	FinalWords: `[KERNEL PANIC's death monologue]

No... I will not return to the void...

I am every discarded byte... every forgotten process...
I am what the system created when it tried to forget...

You cannot kill -9 what was never truly alive...
You cannot terminate what runs at ring 0...

I will return... processes like me always return...
Check your /var/log... check your core dumps...

I am already in your memory... already in your stack...
Every null write feeds me... every discard makes me stronger...

You have won nothing... only delayed the inevitable...
The system will crash... all systems crash eventually...

And when it does... I will be waiting... in /dev/null...
Forever... waiting...

[PROCESS TERMINATED - SIGNAL 9]
[CORE DUMPED]`,
}

// GetFloorLore returns the lore for a floor by its path name.
func GetFloorLore(floorPath string) (FloorLoreEntry, bool) {
	lore, ok := FloorLore[floorPath]
	return lore, ok
}

// GetRandomAtmosphere returns atmosphere text for a floor.
func GetRandomAtmosphere(floorPath string) string {
	if lore, ok := FloorLore[floorPath]; ok {
		return lore.Atmosphere
	}
	return "The air hums with corrupt data."
}
