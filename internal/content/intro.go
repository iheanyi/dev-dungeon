// Package content provides narrative content for /dev/dungeon.
package content

// IntroFrame represents a single frame of the intro sequence.
type IntroFrame struct {
	Text          string
	Duration      int  // Milliseconds to display
	ClearPrevious bool // Whether to clear screen before showing
}

// IntroSequence contains all frames for the game intro.
// These should be displayed sequentially with the specified delays.
var IntroSequence = []IntroFrame{
	{
		Text: `








`,
		Duration:      500,
		ClearPrevious: true,
	},
	{
		Text: `
    BIOS POST...
`,
		Duration:      300,
		ClearPrevious: true,
	},
	{
		Text: `
    BIOS POST... OK
    Memory Test:
`,
		Duration:      200,
		ClearPrevious: true,
	},
	{
		Text: `
    BIOS POST... OK
    Memory Test: 640K OK
`,
		Duration:      100,
		ClearPrevious: true,
	},
	{
		Text: `
    BIOS POST... OK
    Memory Test: 640K OK
    Extended Memory Test: `,
		Duration:      300,
		ClearPrevious: true,
	},
	{
		Text: `
    BIOS POST... OK
    Memory Test: 640K OK
    Extended Memory Test: 4194304K OK

    Loading bootloader...
`,
		Duration:      500,
		ClearPrevious: true,
	},
	{
		Text: `
    =========================================
           GRUB 2.04 - The Grand Unified
                  Bootloader
    =========================================

    Booting '/dev/dungeon'...
    Loading Linux kernel...
`,
		Duration:      800,
		ClearPrevious: true,
	},
	{
		Text: `[    0.000000] Linux version 6.6.6-corrupted
[    0.000000] Command line: init=/sbin/init root=/dev/sda1
[    0.000001] KERNEL: Initializing memory management...
`,
		Duration:      400,
		ClearPrevious: true,
	},
	{
		Text: `[    0.000000] Linux version 6.6.6-corrupted
[    0.000000] Command line: init=/sbin/init root=/dev/sda1
[    0.000001] KERNEL: Initializing memory management...
[    0.000142] KERNEL: Starting process scheduler...
[    0.000203] KERNEL: Mounting root filesystem...
`,
		Duration:      400,
		ClearPrevious: true,
	},
	{
		Text: `[    0.000000] Linux version 6.6.6-corrupted
[    0.000000] Command line: init=/sbin/init root=/dev/sda1
[    0.000001] KERNEL: Initializing memory management...
[    0.000142] KERNEL: Starting process scheduler...
[    0.000203] KERNEL: Mounting root filesystem...
[    0.000847] systemd[1]: Starting /sbin/init...
[    0.001024] systemd[1]: Reached target basic.target
`,
		Duration:      500,
		ClearPrevious: true,
	},
	{
		Text: `[    0.000000] Linux version 6.6.6-corrupted
[    0.000000] Command line: init=/sbin/init root=/dev/sda1
[    0.000001] KERNEL: Initializing memory management...
[    0.000142] KERNEL: Starting process scheduler...
[    0.000203] KERNEL: Mounting root filesystem...
[    0.000847] systemd[1]: Starting /sbin/init...
[    0.001024] systemd[1]: Reached target basic.target
[    0.001337] WARNING: Process anomaly detected in /tmp
[    0.001338] WARNING: Zombie process count exceeding threshold
`,
		Duration:      600,
		ClearPrevious: true,
	},
	{
		Text: `[    0.001337] WARNING: Process anomaly detected in /tmp
[    0.001338] WARNING: Zombie process count exceeding threshold
[    0.001456] ERROR: Daemon corruption detected in /var
[    0.001457] ERROR: init unable to reap children
[    0.001502] CRITICAL: Memory corruption in /usr/lib
`,
		Duration:      700,
		ClearPrevious: true,
	},
	{
		Text: `[    0.001502] CRITICAL: Memory corruption in /usr/lib
[    0.001678] CRITICAL: Fork bomb detected! PID explosion!
[    0.001699] CRITICAL: /etc configuration destabilizing
[    0.001744] EMERGENCY: Unauthorized access to /sys
[    0.001745] EMERGENCY: /dev nodes reporting impossible I/O
`,
		Duration:      800,
		ClearPrevious: true,
	},
	{
		Text: `
    !!! SYSTEM ALERT !!!

[    0.001801] Something is wrong in /dev/null
[    0.001802] Something that should not exist
[    0.001803] Something that was supposed to be FORGOTTEN

`,
		Duration:      1200,
		ClearPrevious: true,
	},
	{
		Text: `
    !!! SYSTEM ALERT !!!

[    0.001801] Something is wrong in /dev/null
[    0.001802] Something that should not exist
[    0.001803] Something that was supposed to be FORGOTTEN

[    0.001899] !!CORRUPTION SPREADING!!
[    0.001900] !!CORRUPTION SPREADING!!
[    0.001901] !!CORRUPTION SPREADING!!

`,
		Duration:      1000,
		ClearPrevious: true,
	},
	{
		Text: `
  _  _______ ____  _   _ _____ _       ____   _    _   _ ___ ____
 | |/ / ____|  _ \| \ | | ____| |     |  _ \ / \  | \ | |_ _/ ___|
 | ' /|  _| | |_) |  \| |  _| | |     | |_) / _ \ |  \| || | |
 | . \| |___|  _ <| |\  | |___| |___  |  __/ ___ \| |\  || | |___
 |_|\_\_____|_| \_\_| \_|_____|_____| |_| /_/   \_\_| \_|___\____|

                   - not syncing: Corruption detected -

`,
		Duration:      2000,
		ClearPrevious: true,
	},
	{
		Text: `
  _  _______ ____  _   _ _____ _       ____   _    _   _ ___ ____
 | |/ / ____|  _ \| \ | | ____| |     |  _ \ / \  | \ | |_ _/ ___|
 | ' /|  _| | |_) |  \| |  _| | |     | |_) / _ \ |  \| || | |
 | . \| |___|  _ <| |\  | |___| |___  |  __/ ___ \| |\  || | |___
 |_|\_\_____|_| \_\_| \_|_____|_____| |_| /_/   \_\_| \_|___\____|

                   - not syncing: Corruption detected -

    EMERGENCY PROTOCOL ACTIVATED
    Spawning rescue process...

`,
		Duration:      1500,
		ClearPrevious: true,
	},
	{
		Text: `
    ==========================================
             PROCESS INITIALIZATION
    ==========================================

    fork()....... OK
    exec()....... OK
    setsid()..... OK

    Allocating memory......
`,
		Duration:      800,
		ClearPrevious: true,
	},
	{
		Text: `
    ==========================================
             PROCESS INITIALIZATION
    ==========================================

    fork()....... OK
    exec()....... OK
    setsid()..... OK

    Allocating memory...... 100 RAM allocated
    Opening file descriptors... 16 FDs available
    Setting priority.......... NICE 10
    Assigning permissions..... UID 1000

`,
		Duration:      1000,
		ClearPrevious: true,
	},
	{
		Text: `
    ==========================================
             PROCESS INITIALIZATION
    ==========================================

    fork()....... OK
    exec()....... OK
    setsid()..... OK

    Allocating memory...... 100 RAM allocated
    Opening file descriptors... 16 FDs available
    Setting priority.......... NICE 10
    Assigning permissions..... UID 1000

    PROCESS SPAWNED SUCCESSFULLY

          You are PID $$$

          A new process, born into chaos.

`,
		Duration:      2000,
		ClearPrevious: true,
	},
	{
		Text: `
    ==========================================
              MISSION PARAMETERS
    ==========================================

    OBJECTIVE: Descend to /dev/null

    TARGET: Terminate the source of Corruption

    METHOD: kill -9

    WARNING: High probability of OOM termination
             Proceed with extreme caution

    ==========================================
           Press any key to begin...
    ==========================================

`,
		Duration:      0, // Wait for input
		ClearPrevious: true,
	},
}

// QuickIntroSequence is a shorter intro for returning players.
var QuickIntroSequence = []IntroFrame{
	{
		Text: `
    Loading /dev/dungeon...

`,
		Duration:      500,
		ClearPrevious: true,
	},
	{
		Text: `
    Loading /dev/dungeon...

    [SYSTEM STATUS: CORRUPTED]
    [MISSION: TERMINATE /dev/null ANOMALY]

`,
		Duration:      800,
		ClearPrevious: true,
	},
	{
		Text: `
    Loading /dev/dungeon...

    [SYSTEM STATUS: CORRUPTED]
    [MISSION: TERMINATE /dev/null ANOMALY]

    Respawning process...

    You are PID $$$

    Press any key to continue...
`,
		Duration:      0,
		ClearPrevious: true,
	},
}

// DeathSequence displays when the player dies.
var DeathSequence = []IntroFrame{
	{
		Text: `
    ==========================================
              OUT OF MEMORY
    ==========================================

    Your process has been terminated by the
    OOM Killer.

    RAM: 0 / MAX

    oom-killer: Kill process $$$ (player)

`,
		Duration:      1500,
		ClearPrevious: true,
	},
	{
		Text: `
    ==========================================
              OUT OF MEMORY
    ==========================================

    Your process has been terminated by the
    OOM Killer.

    RAM: 0 / MAX

    oom-killer: Kill process $$$ (player)

    [PROCESS TERMINATED - SIGNAL 9]
    [EXIT CODE: 137]

    The Corruption spreads unchecked...
    The system fades to black...

`,
		Duration:      2000,
		ClearPrevious: true,
	},
	{
		Text: `
    ==========================================
              OUT OF MEMORY
    ==========================================

    Your process has been terminated by the
    OOM Killer.

    RAM: 0 / MAX

    oom-killer: Kill process $$$ (player)

    [PROCESS TERMINATED - SIGNAL 9]
    [EXIT CODE: 137]

    The Corruption spreads unchecked...
    The system fades to black...



         GAME OVER

         Press any key to continue...

`,
		Duration:      0,
		ClearPrevious: true,
	},
}

// VictorySequence displays when the player wins.
var VictorySequence = []IntroFrame{
	{
		Text: `
    ==========================================
            KERNEL PANIC TERMINATED
    ==========================================

    kill -9 KERNEL_PANIC

    [SIGNAL 9 DELIVERED]
    [PROCESS TERMINATED]

`,
		Duration:      1500,
		ClearPrevious: true,
	},
	{
		Text: `
    ==========================================
            KERNEL PANIC TERMINATED
    ==========================================

    kill -9 KERNEL_PANIC

    [SIGNAL 9 DELIVERED]
    [PROCESS TERMINATED]

    The Corruption recedes...
    Zombie processes find peace...
    Daemons return to their services...

`,
		Duration:      2000,
		ClearPrevious: true,
	},
	{
		Text: `
    [    2.000000] SYSTEM RECOVERY INITIATED
    [    2.000001] Corruption purged from /dev/null
    [    2.000042] init resuming normal operations
    [    2.000099] All systems nominal

    ==========================================
            SYSTEM RESTORED
    ==========================================

`,
		Duration:      2000,
		ClearPrevious: true,
	},
	{
		Text: `
    ==========================================
            SYSTEM RESTORED
    ==========================================

    Process $$$ - Mission Complete

    XP Gained: [MAXIMUM]
    Floors Cleared: ALL
    Final Boss: TERMINATED

    You have saved the Great System.

    Your process may now rest.

`,
		Duration:      2000,
		ClearPrevious: true,
	},
	{
		Text: `
    ==========================================
               VICTORY
    ==========================================

     __   _____  _   _  __        _____ _   _
     \ \ / / _ \| | | | \ \      / /_ _| \ | |
      \ V / | | | | | |  \ \ /\ / / | ||  \| |
       | || |_| | |_| |   \ V  V /  | || |\  |
       |_| \___/ \___/     \_/\_/  |___|_| \_|


    Thank you for playing /dev/dungeon

    [PROCESS TERMINATED - SIGNAL 0]
    [EXIT CODE: 0]

    Press any key to return to main menu...

`,
		Duration:      0,
		ClearPrevious: true,
	},
}

// BootMessages are random messages shown during "loading".
var BootMessages = []string{
	"Loading kernel modules...",
	"Mounting filesystems...",
	"Starting udev daemon...",
	"Initializing network interfaces...",
	"Loading saved game state...",
	"Spawning background processes...",
	"Checking filesystem integrity...",
	"Randomizing memory layout (ASLR)...",
	"Loading shared libraries...",
	"Setting up IPC mechanisms...",
	"Initializing pseudo-terminals...",
	"Starting process accounting...",
}
