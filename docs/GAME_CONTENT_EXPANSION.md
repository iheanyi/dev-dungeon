# /dev/dungeon - Content Expansion Design Document

This document contains comprehensive designs for new content to add depth and replayability to /dev/dungeon. All content follows the Unix/hacker theme and integrates with existing systems.

---

## Table of Contents

1. [Floor Variants](#1-floor-variants)
2. [New Enemies](#2-new-enemies)
3. [New Character Classes](#3-new-character-classes)
4. [New Items](#4-new-items)
5. [Side Dungeons](#5-side-dungeons)
6. [Endless Mode](#6-endless-mode)
7. [Boss Encounters](#7-boss-encounters)

---

## 1. Floor Variants

Each main floor type now has 2-3 variants with distinct enemy compositions, environmental hazards, and loot tables. When entering a floor, one variant is randomly selected (or influenced by player actions).

### /home Variants

#### /home/guest (Easy)
**Theme:** A restricted guest account with limited permissions
**Enemy Composition:** tutorial_bug, lint_error, deprecated_warning
**Special Mechanic:** Cannot access certain rooms without finding guest credentials
**Environmental Hazard:** None (safe zone)
**Loot Modifier:** Common items only, guaranteed starter weapon
**Flavor:** "Welcome, guest. Your session will expire in... never mind, time has no meaning here."

#### /home/.hidden (Easy-Medium)
**Theme:** Hidden dotfiles containing secrets and configurations
**Enemy Composition:** dotfile_watcher, config_parser, symlink_trap
**Special Mechanic:** Map is initially invisible; must find `.bashrc` to reveal layout
**Environmental Hazard:** Symlink loops can teleport player randomly
**Loot Modifier:** Higher chance of utility items, hidden caches
**Flavor:** "Files that begin with a dot are hidden for a reason. Some secrets are better left undiscovered."

#### /home/root (Medium)
**Theme:** Accidentally stumbled into root's home directory
**Enemy Composition:** permission_denied, auth_log, sudo_timer
**Special Mechanic:** High UID enemies patrol; stealth is recommended
**Environmental Hazard:** Failed actions alert nearby enemies
**Loot Modifier:** Rare items possible but heavily guarded
**Flavor:** "You shouldn't be here. Neither should the things that live here now."

---

### /tmp Variants

#### /tmp/cache (Easy-Medium)
**Theme:** Stale cached data accumulating corruption
**Enemy Composition:** stale_cache, cache_miss, memory_leak
**Special Mechanic:** Enemies grow stronger the longer you stay on the floor
**Environmental Hazard:** Cache invalidation zones deal MEM damage over time
**Loot Modifier:** Consumables have +1 stack size
**Flavor:** "There are only two hard things in computer science: cache invalidation and... wait, what were we talking about?"

#### /tmp/orphans (Medium)
**Theme:** Abandoned child processes with no parent
**Enemy Composition:** orphan_process, init_adopter, process_reaper
**Special Mechanic:** Killing an enemy may spawn its "children" (weaker versions)
**Environmental Hazard:** Orphan swarms trigger every 15 turns
**Loot Modifier:** Bonus exit codes on completion
**Flavor:** "When a parent process dies, init adopts its children. But something else got here first."

#### /tmp/session (Easy-Medium)
**Theme:** Active user sessions competing for resources
**Enemy Composition:** idle_session, tty_hijacker, screen_detach
**Special Mechanic:** Other "players" (NPCs) compete for loot and stairs
**Environmental Hazard:** Sessions can timeout, removing safe zones
**Loot Modifier:** Race condition - first to a chest gets better loot
**Flavor:** "Last login: 3 years ago. Session still active."

---

### /var Variants

#### /var/log (Medium)
**Theme:** Endless logging that records everything
**Enemy Composition:** log_rotate, syslog_daemon, journald_bloat
**Special Mechanic:** Your actions are "logged" - enemies know your last 5 moves
**Environmental Hazard:** Log overflow zones slow movement
**Loot Modifier:** Information items (maps, enemy stats) more common
**Flavor:** "Everything you do here is being recorded. Everything you've ever done has already been written."

#### /var/spool (Medium)
**Theme:** Print queues and mail spools clogged with ancient jobs
**Enemy Composition:** print_job, mail_bouncer, queue_stuck, cron_mail
**Special Mechanic:** Enemies form queues - must defeat in order (FIFO)
**Environmental Hazard:** Priority inversion can shuffle enemy order
**Loot Modifier:** Queued items - retrieve later at /home
**Flavor:** "Your print job is number 847,293 in the queue. Please do not turn off the printer."

#### /var/run (Medium-Hard)
**Theme:** Runtime data and PID files of active daemons
**Enemy Composition:** stale_pid, runaway_daemon, lock_file, socket_listener
**Special Mechanic:** Enemies respawn unless their "PID file" is deleted
**Environmental Hazard:** Lock contentions freeze player for 1 turn
**Loot Modifier:** Service tokens more common
**Flavor:** "Every daemon leaves a PID file. Kill the file, kill the daemon. Simple. In theory."

---

### /etc Variants

#### /etc/shadow (Hard)
**Theme:** Encrypted password hashes and authentication data
**Enemy Composition:** hash_collision, salt_miner, rainbow_table, brute_forcer
**Special Mechanic:** Enemies have "encrypted" health - must crack to see true HP
**Environmental Hazard:** Wrong passwords trigger alarm (spawn enemies)
**Loot Modifier:** Authentication items, UID boosters
**Flavor:** "The shadow file contains hashed secrets. Something has been trying to crack them for centuries."

#### /etc/cron.d (Medium-Hard)
**Theme:** Scheduled tasks running at chaotic intervals
**Enemy Composition:** cron_expression, at_job, anacron_delay, timer_unit
**Special Mechanic:** Enemies only active at certain "times" (turn counts)
**Environmental Hazard:** Time skips can occur (5-10 turns pass instantly)
**Loot Modifier:** Time-based buffs and scheduling items
**Flavor:** "0 * * * * /usr/bin/nightmare --repeat"

#### /etc/init.d (Hard)
**Theme:** Ancient SysV init scripts awakening
**Enemy Composition:** init_script, runlevel_shift, dependency_hell, service_loop
**Special Mechanic:** Killing enemies in wrong order causes respawn chain
**Environmental Hazard:** Runlevel changes alter floor layout
**Loot Modifier:** Process control items, init tokens
**Flavor:** "These scripts predate systemd. They predate sanity. They do not die easily."

---

### /usr Variants

#### /usr/bin (Hard)
**Theme:** User binaries gone rogue
**Enemy Composition:** deprecated_binary, setuid_binary, wrapper_script, busybox_hydra
**Special Mechanic:** Enemies can "exec" into other enemy types
**Environmental Hazard:** Binary incompatibility zones (random stat debuffs)
**Loot Modifier:** Weapon drops more common
**Flavor:** "Every binary here was installed by someone, for some reason. Most of those reasons are lost to time."

#### /usr/share (Medium-Hard)
**Theme:** Shared resources and documentation
**Enemy Composition:** locale_error, missing_font, broken_link, doc_rot
**Special Mechanic:** Information overload - too many items causes confusion debuff
**Environmental Hazard:** Locale changes scramble UI temporarily
**Loot Modifier:** Man pages and skill items more common
**Flavor:** "Shared knowledge is power. Shared corruption is also power, just the wrong kind."

#### /usr/local (Very Hard)
**Theme:** Locally compiled software, untested and dangerous
**Enemy Composition:** unvetted_build, compile_error, undefined_symbol, abi_break
**Special Mechanic:** Enemy behavior is "undefined" - random attack patterns
**Environmental Hazard:** Segfault zones cause instant damage
**Loot Modifier:** Powerful but unstable items
**Flavor:** "You compiled this yourself. You have no one to blame but yourself."

---

### /sys Variants

#### /sys/kernel (Very Hard)
**Theme:** Direct kernel interface, reality is unstable
**Enemy Composition:** kernel_thread, interrupt_handler, softirq, workqueue
**Special Mechanic:** Player NICE stat fluctuates wildly
**Environmental Hazard:** Preemption can interrupt player actions
**Loot Modifier:** Kernel-level abilities and system calls
**Flavor:** "You're not supposed to be able to see this. The kernel is not supposed to be visible."

#### /sys/devices (Very Hard)
**Theme:** Hardware abstraction layer breaking down
**Enemy Composition:** null_device, full_device, zero_flood, random_chaos
**Special Mechanic:** Device files have special interactions (read/write mechanics)
**Environmental Hazard:** /dev/random zones cause unpredictable effects
**Loot Modifier:** Device control items
**Flavor:** "Everything is a file. Even the things that shouldn't be files. Especially those."

#### /sys/power (Brutal)
**Theme:** Power management in a system that refuses to sleep
**Enemy Composition:** suspend_blocker, wakelock_hog, acpi_event, thermal_throttle
**Special Mechanic:** System tries to "suspend" - must stay active to survive
**Environmental Hazard:** Power zones drain all resources slowly
**Loot Modifier:** Stamina and efficiency items
**Flavor:** "The system is trying to sleep. Something is keeping it awake. Something hungry."

---

### /dev Variants

#### /dev/tty (Brutal)
**Theme:** Terminal devices with minds of their own
**Enemy Composition:** pty_master, tty_slave, line_discipline, terminal_escape
**Special Mechanic:** Limited visibility - enemies can attack from off-screen
**Environmental Hazard:** Escape sequences can disorient (swap controls)
**Loot Modifier:** Terminal control items
**Flavor:** "The terminal is a window into the system. Something is looking back through it."

#### /dev/shm (Brutal)
**Theme:** Shared memory filled with corrupted data
**Enemy Composition:** memory_corruption, race_condition, use_after_free, double_free
**Special Mechanic:** Memory attacks can "corrupt" inventory items
**Environmental Hazard:** Data races cause duplicate/missing enemies
**Loot Modifier:** Memory manipulation items
**Flavor:** "Shared memory means shared nightmares. What one process writes, all processes read."

#### /dev/mapper (Brutal)
**Theme:** Logical volume management gone wrong
**Enemy Composition:** lvm_vg, thin_pool, snapshot_ghost, stripe_corruption
**Special Mechanic:** Floor layout shifts dynamically as "volumes" resize
**Environmental Hazard:** Volume collapse can trap player
**Loot Modifier:** Storage and space items
**Flavor:** "The mapper organizes chaos into order. But chaos is organized now, and it has opinions."

---

## 2. New Enemies

### Tier 1 (Easy - /home, /tmp)

#### lint_error `!`
**Stats:** PID 15 | CPU 3 | MEM 0 | NICE 15
**Behavior:** Passive until player makes a combat mistake (miss or fumble)
**Special Ability:** *Code Review* - Retaliates with double damage on player misses
**Drops:** Style tokens, minor consumables
**Flavor:** "Warning: variable 'sanity' declared but never used."

#### deprecated_warning `~`
**Stats:** PID 20 | CPU 5 | MEM 5 | NICE 18
**Behavior:** Slowly follows player, leaves warning trails
**Special Ability:** *Sunset Timer* - Grows stronger each floor (becomes deprecated_error)
**Drops:** Legacy items, compatibility patches
**Flavor:** "This enemy will be removed in a future version. That version never comes."

#### null_pointer `0`
**Stats:** PID 1 | CPU 25 | MEM 0 | NICE 8
**Behavior:** Erratic teleportation, fragile but dangerous
**Special Ability:** *Dereference* - Instantly kills on contact but dies in one hit
**Drops:** Null fragments, pointer items
**Flavor:** "It points to nothing. Nothing is very, very angry."

---

### Tier 2 (Medium - /var, /etc)

#### memory_leak `%`
**Stats:** PID 40 | CPU 8 | MEM 30 | NICE 12
**Behavior:** Steadily drains player MEM while alive
**Special Ability:** *Heap Growth* - Gains +5 PID each turn until killed
**Drops:** Memory fragments, MEM potions
**Flavor:** "It started as a few bytes. It's been growing for years."

#### race_condition `&`
**Stats:** PID 35 | CPU 15 | MEM 10 | NICE 5
**Behavior:** Alternates between two positions each turn
**Special Ability:** *TOCTOU* - 50% chance to dodge any attack (check vs use)
**Drops:** Synchronization primitives, timing crystals
**Flavor:** "It exists in two states. Neither state is good for you."

#### permission_denied `#`
**Stats:** PID 50 | CPU 12 | MEM 15 | NICE 10
**Behavior:** Blocks paths, immune to low UID players
**Special Ability:** *Access Control* - Cannot be damaged unless player UID > enemy UID
**Drops:** UID shards, permission tokens
**Flavor:** "Permission denied. Permission has always been denied."

#### log_rotate `@`
**Stats:** PID 60 | CPU 10 | MEM 20 | NICE 14
**Behavior:** Circular patrol pattern, records player position
**Special Ability:** *Archive* - On death, spawns compressed version with 50% stats
**Drops:** Log fragments, history items
**Flavor:** "It remembers everything. Even the things you deleted."

#### cron_expression `*`
**Stats:** PID 45 | CPU 18 | MEM 8 | NICE 0 (acts on specific turns)
**Behavior:** Only active on turns divisible by 5
**Special Ability:** *Scheduled Execution* - Triple damage on active turns, dormant otherwise
**Drops:** Schedule tokens, timing items
**Flavor:** "*/5 * * * * /bin/kill $YOU"

---

### Tier 3 (Hard - /usr, /sys)

#### dependency_hell `$`
**Stats:** PID 80 | CPU 20 | MEM 25 | NICE 9
**Behavior:** Spawns with 1-3 linked enemies; killing one heals others
**Special Ability:** *Circular Dependency* - All linked enemies must die in same turn or respawn
**Drops:** Dependency graphs, resolution items
**Flavor:** "Package A requires B. Package B requires C. Package C requires A. Good luck."

#### buffer_overflow `[`
**Stats:** PID 70 | CPU 30 | MEM 40 | NICE 7
**Behavior:** Charges in straight lines, damages everything in path
**Special Ability:** *Stack Smash* - Can overwrite player buffs with debuffs
**Drops:** Buffer items, overflow protection
**Flavor:** "It doesn't care about boundaries. Boundaries don't exist anymore."

#### use_after_free `]`
**Stats:** PID 55 | CPU 35 | MEM 20 | NICE 6
**Behavior:** "Dies" but remains active as ghost until turn ends
**Special Ability:** *Dangling Pointer* - Can attack after "death" if player steps on its space
**Drops:** Memory safety items, free() tokens
**Flavor:** "You freed it. You should have checked if something was still using it."

#### interrupt_handler `^`
**Stats:** PID 65 | CPU 25 | MEM 30 | NICE 1
**Behavior:** Interrupts player actions, forcing them to repeat
**Special Ability:** *IRQ Storm* - Can interrupt the same action multiple times
**Drops:** Handler masks, IRQ tokens
**Flavor:** "It has priority. It always has priority. Your priority is irrelevant."

#### undefined_behavior `?`
**Stats:** PID ??? | CPU ??? | MEM ??? | NICE ???
**Behavior:** All stats randomized each turn (10-100 range)
**Special Ability:** *Nasal Demons* - Random effect on each attack (heal, damage, buff, debuff)
**Drops:** Chaos items, random loot
**Flavor:** "The behavior is undefined. The pain is very defined."

---

### Tier 4 (Brutal - /dev)

#### kernel_oops `K`
**Stats:** PID 120 | CPU 40 | MEM 50 | NICE 3
**Behavior:** Causes system-wide effects when damaged
**Special Ability:** *Taint Flag* - Each hit adds a "taint"; 5 taints trigger floor-wide damage
**Drops:** Kernel fragments, oops dumps
**Flavor:** "Oops. Oops. Oops. It keeps saying that. It doesn't mean it."

#### hypervisor `H`
**Stats:** PID 150 | CPU 35 | MEM 60 | NICE 2
**Behavior:** Creates "virtual" copies of other enemies
**Special Ability:** *VM Escape* - Virtual copies can become real if not killed in 3 turns
**Drops:** Virtualization tokens, escape keys
**Flavor:** "It manages realities. Your reality is up for management."

#### firmware_blob `F`
**Stats:** PID 100 | CPU 45 | MEM 40 | NICE 8
**Behavior:** Slow but immune to most attacks
**Special Ability:** *Binary Only* - Can only be damaged by specific item types
**Drops:** Firmware shards, binary keys
**Flavor:** "No source code. No documentation. No mercy."

---

### Special/Rare Enemies

#### bitcoin_miner `B`
**Stats:** PID 200 | CPU 80 | MEM 100 | NICE 20
**Behavior:** Extremely slow but drains massive resources from nearby
**Special Ability:** *Proof of Work* - Immune until player "solves" puzzle (specific input sequence)
**Drops:** Crypto tokens (high value currency)
**Spawn Condition:** Rare spawn on /usr and below
**Flavor:** "It's using 100% of your CPU. It's been using 100% of everyone's CPU."

#### blockchain `C`
**Stats:** PID 300 | CPU 20 | MEM 150 | NICE 15
**Behavior:** Creates "blocks" that must be cleared to damage it
**Special Ability:** *Immutable* - Cannot be damaged until all blocks (3-5) are destroyed
**Drops:** Chain links, immutability tokens
**Spawn Condition:** Appears if bitcoin_miner is killed
**Flavor:** "The chain remembers. The chain is forever. The chain is also blocking the exit."

---

## 3. New Character Classes

### systemd (Unlockable)

**Theme:** Modern init system with parallel execution
**Starting Stats:**
- PID: 100
- CPU: 15
- MEM: 25
- NICE: 8
- UID: 1

**Passive Ability - *Parallel Execution*:** Can queue 2 actions per turn, but second action has 50% effectiveness. Actions execute simultaneously.

**Active Ability - *Dependency Resolution* (20 MEM):** Target enemy cannot act until specified other enemy is killed. Creates combat puzzles.

**Special Mechanic - *Unit Files*:** Collects "unit files" from enemies. Can activate units for temporary buffs (oneshot effects).

**Unlock Condition:** Defeat 10 enemies of each init script type (init_script, runlevel_shift, dependency_hell) in a single run.

**Playstyle:** Complex action economy, planning-focused. Rewards players who think ahead and manage multiple threats.

**Flavor:** "PID 1. The first process. The last line of defense. You are the init, and all processes are your children."

---

### grep (Unlockable)

**Theme:** Pattern matching and information warfare
**Starting Stats:**
- PID: 70
- CPU: 10
- MEM: 40
- NICE: 12
- UID: 0

**Passive Ability - *Regular Expression*:** Attacks that match enemy "patterns" (specific stat thresholds) deal 2x damage. Patterns revealed on enemy inspection.

**Active Ability - *Recursive Search* (30 MEM):** Reveals all enemies on floor, their stats, and any hidden items. Information persists until floor change.

**Special Mechanic - *Pattern Library*:** Builds library of enemy "patterns." Known patterns provide permanent +10% damage against that enemy type.

**Unlock Condition:** Use grep scroll on every floor in a single winning run.

**Playstyle:** Information-focused. Weaker in direct combat but excels at preparation and targeting weaknesses.

**Flavor:** "You search for patterns in chaos. Sometimes you find them. Sometimes they find you."

---

### awk (Unlockable)

**Theme:** Text processing and transformation
**Starting Stats:**
- PID: 80
- CPU: 12
- MEM: 35
- NICE: 10
- UID: 0

**Passive Ability - *Field Separator*:** Attacks split enemy stats into "fields." Can target individual stats (CPU, MEM) instead of PID.

**Active Ability - *Transform* (25 MEM):** Convert enemy stat points between types. Remove 20 CPU to heal 20 PID, etc.

**Special Mechanic - *Print Statement*:** Defeating enemies generates "output" text. Collecting specific outputs unlocks bonus abilities.

**Unlock Condition:** Defeat the kernel_panic using only basic attacks (no abilities or items).

**Playstyle:** Tactical stat manipulation. Excels at weakening strong enemies and converting resources.

**Flavor:** "BEGIN { survival = 0 } /dungeon/ { survival++ } END { print survival }"

---

### ssh (Unlockable)

**Theme:** Secure shell, remote execution, tunneling
**Starting Stats:**
- PID: 90
- CPU: 14
- MEM: 30
- NICE: 6
- UID: 2

**Passive Ability - *Tunneling*:** Can "tunnel" through walls to adjacent rooms. Limited uses per floor (3).

**Active Ability - *Remote Execution* (35 MEM):** Execute an action on target enemy's "host" - damage affects all enemies of same type on floor.

**Special Mechanic - *Port Forwarding*:** Create "portals" between rooms. Can set up escape routes or ambush points.

**Unlock Condition:** Complete a run without ever fleeing from combat.

**Playstyle:** Mobility and area control. Excels at positioning and affecting multiple enemies.

**Flavor:** "The secure shell protects you. The secure shell connects everything. Everything connected can be reached."

---

### make (Unlockable - Secret Class)

**Theme:** Build system, dependency management, automation
**Starting Stats:**
- PID: 60
- CPU: 8
- MEM: 50
- NICE: 14
- UID: 0

**Passive Ability - *Makefile*:** Automatically executes optimal action based on game state (AI-assisted). Player can override.

**Active Ability - *Rebuild* (40 MEM):** Fully restore one piece of equipment or "rebuild" a dead ally (if multiplayer) with 50% stats.

**Special Mechanic - *Build Targets*:** Collect "source files" from enemies. Combine sources to build powerful items mid-run.

**Unlock Condition:** Complete endless mode to depth 50.

**Playstyle:** Automation and crafting. Less direct control but powerful emergent combinations.

**Flavor:** "make: 'victory' is up to date. make: Nothing to be done for 'survival'."

---

## 4. New Items

### Weapons

#### segfault_handler `!`
**Type:** Weapon (Melee)
**Rarity:** Rare
**Stats:** +18 CPU, +5% critical chance
**Special Effect:** Critical hits cause "segfault" status - enemy skips next turn and takes 10 damage
**Flavor:** "It doesn't prevent segfaults. It just makes them someone else's problem."

#### pipe_chain `|`
**Type:** Weapon (Ranged)
**Rarity:** Uncommon
**Stats:** +12 CPU
**Special Effect:** Damage increases by +3 for each consecutive hit (resets on miss). Max +15 bonus.
**Flavor:** "The output of one strike becomes the input of the next."

#### regex_blade `/`
**Type:** Weapon (Melee)
**Rarity:** Rare
**Stats:** +15 CPU, -2 NICE
**Special Effect:** Deals bonus damage based on enemy name length (+1 per character)
**Flavor:** "Some people, when confronted with a problem, think 'I know, I'll use regular expressions.' Now they have two problems and this sword."

#### fork_bomb_defused `:`
**Type:** Weapon (Melee)
**Rarity:** Epic
**Stats:** +25 CPU
**Special Effect:** On kill, 20% chance to spawn a friendly "fork" that attacks enemies for 3 turns
**Flavor:** ":(){ :|:& };: - but controlled. Barely."

#### null_terminator `\0`
**Type:** Weapon (Melee)
**Rarity:** Legendary
**Stats:** +30 CPU
**Special Effect:** Instantly kills enemies below 10% PID. "Terminates" their string.
**Flavor:** "Every string must end. You are the ending."

---

### Armor

#### firewall `#`
**Type:** Armor (Body)
**Rarity:** Uncommon
**Stats:** +15 PID, +10% block chance
**Special Effect:** Blocks all damage from enemies you haven't attacked yet this combat
**Flavor:** "Default deny. Everything is denied until proven friendly."

#### sandbox `[ ]`
**Type:** Armor (Body)
**Rarity:** Rare
**Stats:** +20 PID
**Special Effect:** Take 50% damage from first hit of each combat. "Contained" execution.
**Flavor:** "The sandbox contains threats. The sandbox contains you. Same thing, really."

#### encryption_layer `{ }`
**Type:** Armor (Body)
**Rarity:** Epic
**Stats:** +25 PID, +15 MEM
**Special Effect:** 25% chance to completely negate any hit. Enemy attack is "encrypted" (blocked).
**Flavor:** "They can't hurt what they can't read."

#### kernel_module `.ko`
**Type:** Armor (Accessory)
**Rarity:** Legendary
**Stats:** +10 all stats
**Special Effect:** Gain benefits of the current floor's "kernel" - different buff per floor type
**Flavor:** "Load the module. Become part of the kernel. Lose yourself in the system."

---

### Consumables

#### core_dump `core`
**Type:** Consumable
**Rarity:** Common
**Effect:** Reveals all enemy stats and abilities for current floor
**Flavor:** "A snapshot of the moment everything went wrong. Very informative."

#### reboot_script `#!/`
**Type:** Consumable
**Rarity:** Uncommon
**Effect:** Fully restores PID and MEM. Clears all status effects (good and bad).
**Flavor:** "Have you tried turning yourself off and on again?"

#### stack_trace `bt`
**Type:** Consumable
**Rarity:** Rare
**Effect:** Rewind last 3 actions (yours and enemies). Time manipulation.
**Flavor:** "The trace shows where you went wrong. Now you can go right."

#### memory_dump `dd`
**Type:** Consumable
**Rarity:** Uncommon
**Effect:** Copies a random buff from target enemy to yourself
**Flavor:** "dd if=/enemy/buff of=/self/buff bs=1M"

#### cache_flush `sync`
**Type:** Consumable
**Rarity:** Common
**Effect:** Removes all debuffs. Slight MEM cost (5).
**Flavor:** "Flush the buffers. Clear the cache. Start fresh."

#### fork() `f()`
**Type:** Consumable
**Rarity:** Epic
**Effect:** Create a temporary clone that mimics your last action for 5 turns
**Flavor:** "You are now two processes. Try to stay synchronized."

---

### Utilities

#### symbolic_link `->`
**Type:** Utility
**Rarity:** Uncommon
**Effect:** Link two rooms together. Can traverse instantly. Single use per floor.
**Flavor:** "It points to somewhere else. Somewhere else points back."

#### environment_variable `$VAR`
**Type:** Utility
**Rarity:** Rare
**Effect:** Store one item in "environment." Can be retrieved on any future floor.
**Flavor:** "export SALVATION='this item'. It will be there when you need it."

#### process_priority `nice`
**Type:** Utility
**Rarity:** Uncommon
**Effect:** Permanently modify your NICE by +/- 3 (choose on use). Can only use once per run.
**Flavor:** "Be nicer, be slower. Be meaner, be faster. Choose."

#### shell_expansion `{a,b}`
**Type:** Utility
**Rarity:** Rare
**Effect:** Duplicate a consumable item (one use only). Cannot duplicate itself.
**Flavor:** "touch {item,item_copy}. Shell expansion is powerful and confusing."

#### background_job `&`
**Type:** Utility
**Rarity:** Epic
**Effect:** Queue an action to execute automatically in 3 turns. Plan ahead.
**Flavor:** "Run it in the background. Hope it finishes before you do."

---

## 5. Side Dungeons

### /proc - The Process Filesystem

**Access:** Discovered via secret room in /var or /etc. Requires finding `/proc/self` item.

**Theme:** A surreal reflection of the player's own process state

**Structure:** 3 floors, mirrored/recursive layout

#### /proc/self
**Description:** A mirror of your own process. Enemies are corrupted versions of yourself.
**Enemies:** shadow_self, memory_image, thread_clone
**Special Mechanic:** Damage dealt to enemies affects your own stats (risk/reward)
**Boss:** pid_max - Your maximum potential, corrupted
**Rewards:** Permanent stat insight (see hidden stats), self_reference item

#### /proc/meminfo
**Description:** Navigate the memory map of the system
**Enemies:** page_fault, swap_thrash, oom_killer, memory_balloon
**Special Mechanic:** MEM-focused combat. High MEM cost abilities but powerful effects.
**Boss:** memory_pressure - A crushing force of resource exhaustion
**Rewards:** +20 permanent MEM, memory manipulation abilities

#### /proc/interrupts
**Description:** The interrupt table made manifest
**Enemies:** timer_interrupt, keyboard_irq, network_irq, spurious_interrupt
**Special Mechanic:** Time-based puzzle combat. Must act during correct interrupt windows.
**Boss:** interrupt_storm - All interrupts firing simultaneously
**Rewards:** NICE reduction (permanent), interrupt immunity item

**Unique Loot Table:**
- `/proc/cmdline`: Reveals game seed and hidden mechanics
- `/proc/uptime`: Shows total playtime, grants bonus based on efficiency
- `/proc/loadavg`: Scales difficulty dynamically based on player skill

---

### /mnt - The Mounted Drives

**Access:** Found by discovering an unmounted device in /dev. Requires mount command item.

**Theme:** External filesystems with alien logic

**Structure:** 4 "mounted" areas, each with different rules

#### /mnt/usb
**Description:** A flash drive from another system, containing foreign processes
**Enemies:** autorun_virus, driver_conflict, filesystem_corruption, usb_worm
**Special Mechanic:** Enemies have "foreign" abilities not seen elsewhere
**Unique Hazard:** Eject danger - floor can "unmount" with 10-turn warning
**Rewards:** Portable items (lighter weight), external_backup item

#### /mnt/network
**Description:** A network share, enemies from remote systems
**Enemies:** packet_loss, latency_spike, connection_reset, man_in_middle
**Special Mechanic:** "Lag" affects all actions - results delayed by 1-2 turns
**Unique Hazard:** Connection drops can split party (multiplayer) or strand player
**Rewards:** Network abilities, remote_execute item

#### /mnt/encrypted
**Description:** An encrypted volume with scrambled reality
**Enemies:** encrypted_block, key_derivation, cipher_chain, entropy_pool
**Special Mechanic:** All information scrambled until "key" is found
**Unique Hazard:** Wrong decryption attempts spawn additional enemies
**Boss:** cryptographic_lock - Must solve cipher puzzle during combat
**Rewards:** Encryption abilities, unbreakable items

#### /mnt/iso
**Description:** A frozen snapshot, read-only reality
**Enemies:** immutable_process, snapshot_ghost, verification_error
**Special Mechanic:** Cannot modify environment, must work around obstacles
**Unique Hazard:** No healing allowed - all changes are "read-only"
**Rewards:** Snapshot ability (save state), immutable_armor

**Unique Loot Table:**
- `mount_point`: Create temporary safe zones
- `fstab_entry`: Permanent mount option, always access one /mnt area
- `umount_scroll`: Instantly banish any single enemy (unmount it)

---

### /opt - Optional Packages

**Access:** Purchase access from special vendor for high exit code cost, or find ancient_package_manager item

**Theme:** Third-party software of varying quality and danger

**Structure:** Procedurally generated based on "packages" player has encountered

#### /opt/proprietary
**Description:** Closed-source software with hidden behaviors
**Enemies:** license_violation, drm_guardian, telemetry_collector, black_box
**Special Mechanic:** Enemy abilities are hidden until encountered multiple times
**Unique Hazard:** EULA traps - must accept to proceed, with random penalties
**Rewards:** Powerful but unpredictable items, proprietary_key

#### /opt/experimental
**Description:** Bleeding-edge software, powerful but unstable
**Enemies:** beta_feature, regression_bug, breaking_change, untested_code
**Special Mechanic:** All abilities (player and enemy) have 20% chance of "undefined behavior"
**Unique Hazard:** Frequent crashes (save points between rooms)
**Rewards:** Experimental abilities, cutting_edge items

#### /opt/abandoned
**Description:** Unmaintained packages left to rot
**Enemies:** bit_rot, abandoned_feature, security_vulnerability, forgotten_code
**Special Mechanic:** Enemies grow weaker over "time" (turns) but floor grows darker
**Boss:** unmaintained_dependency - A massive hulk of technical debt
**Rewards:** Vintage items (old but reliable), legacy_power

**Unique Loot Table:**
- `package_manager`: Combine items to create new ones
- `version_pin`: Lock an item's stats permanently (no degradation)
- `source_code`: Understand any item completely (full stat reveal)

---

## 6. Endless Mode

### Core Design

**Name:** `/dev/infinity` or "The Recursive Descent"

**Access:** Unlocked after defeating kernel_panic. Start from special node in /dev/null.

**Theme:** The system descending into infinite recursion, reality breaking down

### Structure

**Depth System:** Floors numbered 1 to infinity. Every 10 floors is a "checkpoint" where player can extract with partial rewards.

**Floor Generation Rules:**
- Depths 1-10: Random selection from all standard floor variants (remixed)
- Depths 11-20: Variants start combining (e.g., /var/log meets /etc/shadow)
- Depths 21-30: Corrupted variants (familiar layouts with wrong enemies)
- Depths 31-40: Procedural variants (completely new combinations)
- Depths 41-50: Reality breakdown (multiple floor rules active simultaneously)
- Depths 51+: Pure chaos (all rules optional, anything can happen)

**Scaling Formula:**
```
enemy_stat_multiplier = 1.0 + (depth * 0.05)
enemy_count_multiplier = 1.0 + (depth * 0.02)
loot_quality_bonus = depth * 2
```

### Endless-Only Enemies

#### recursion_overflow `R`
**Stats (Base):** PID 100 | CPU 30 | MEM 30 | NICE 5
**Behavior:** Creates smaller copies of itself when damaged
**Special:** Copies create copies. Must kill all within 5 turns or they reform.
**First Appears:** Depth 15
**Flavor:** "f(x) = f(x). The base case was never defined."

#### stack_exhaust `S`
**Stats (Base):** PID 80 | CPU 40 | MEM 60 | NICE 8
**Behavior:** Each turn, gains +10 to one random stat, loses -10 from another
**Special:** On death, releases all accumulated stats as AoE damage
**First Appears:** Depth 25
**Flavor:** "The stack grows. The stack must stop growing. The stack doesn't care."

#### heap_spray `Y`
**Stats (Base):** PID 60 | CPU 20 | MEM 80 | NICE 12
**Behavior:** Spawns "heap fragments" that block movement and deal chip damage
**Special:** Fragments persist after enemy death, must be cleared manually
**First Appears:** Depth 30
**Flavor:** "It fills all available memory. Your memory. Everyone's memory."

#### entropy_death `E`
**Stats (Base):** PID 150 | CPU 50 | MEM 50 | NICE 3
**Behavior:** All randomness near it becomes deterministic (predictable but unavoidable)
**Special:** Removes all random elements from combat - pure skill check
**First Appears:** Depth 40
**Flavor:** "Randomness is dead. What remains is certainty. Certain doom."

#### void_pointer `V`
**Stats (Base):** PID 1 | CPU 100 | MEM 0 | NICE 1
**Behavior:** Exists in quantum state - may or may not be real until observed
**Special:** 50% chance to not exist when attacked. Attacks always hit if it exists.
**First Appears:** Depth 50
**Flavor:** "void* ptr = anything; everything = ptr; // This is fine."

### Endless Mode Rewards

**Per-Depth Rewards:**
- Every depth: Guaranteed loot drop
- Every 5 depths: Rare item or ability
- Every 10 depths: Checkpoint shrine (can extract) + legendary item
- Every 25 depths: Unique endless-only item

**Extract Rewards:**
- Depth 10: "Deep Runner" achievement, +50 exit codes
- Depth 25: "Stack Survivor" achievement, +200 exit codes, unlock make class hint
- Depth 50: "Heap Master" achievement, +500 exit codes, unlock make class
- Depth 100: "Recursion Resistant" achievement, +2000 exit codes, secret ending

**Leaderboard Tracking:**
- Deepest depth reached
- Fastest to depth 50
- Most enemies killed in single run
- Highest single-hit damage

---

## 7. Boss Encounters

### Mini-Bosses

#### init_zero (Floor 2-3, /tmp or /var)

**Name:** init_zero - The First Process Gone Wrong
**Glyph:** `I`
**Stats:** PID 150 | CPU 25 | MEM 40 | NICE 5 | UID 0

**Phase 1 (100-50% PID):**
- Standard attacks, spawns child processes (1-2 weak enemies per turn)
- "Fork" ability: Creates weaker clone with 25% stats

**Phase 2 (50-0% PID):**
- Attempts to "adopt" player as child process (charm-like status)
- "Reap" ability: Instantly kills any enemy on field and heals for their PID
- Enrage: All spawned children gain +50% stats

**Arena:** Circular room with process table in center. Table shows spawn countdown.

**Strategy:** Kill children quickly to prevent overwhelming numbers. Save burst for phase 2.

**Drops:**
- init_token (guaranteed): Unlocks systemd class progress
- orphan_adopter (25%): Passive - defeated enemies have 10% chance to fight for you for 3 turns

**Flavor:** "In the beginning, there was init. And init spawned processes. And the processes were good. Then init went wrong."

---

#### cron_daemon (Floor 3-4, /var/spool or /etc/cron.d)

**Name:** cron_daemon - The Eternal Scheduler
**Glyph:** `C`
**Stats:** PID 180 | CPU 30 | MEM 50 | NICE 0 (variable) | UID 0

**Mechanic:** Turn counter determines all boss behavior. Pattern repeats every 10 turns.

**Turn Pattern:**
- Turn 1, 6: Heavy attack (40 damage)
- Turn 2, 7: Spawn adds (schedule_entry enemies)
- Turn 3, 8: Shield (immune to damage)
- Turn 4, 9: AoE attack (20 damage to all)
- Turn 5, 10: Vulnerable (takes 2x damage)

**Special Ability - *Crontab Edit*:** At 50% PID, pattern scrambles. Must learn new pattern.

**Arena:** Clock-themed room with visible turn counter. Floor tiles light up indicating next attack zone.

**Strategy:** Memorize pattern, maximize damage on vulnerable turns, defend on heavy attack turns.

**Drops:**
- cron_expression_book (guaranteed): Learn enemy attack patterns faster
- time_crystal (50%): Consumable - skip enemy turn once per combat
- scheduled_strike (25%): Weapon - attacks deal bonus damage on specific turns

**Flavor:** "It runs on schedule. It has always run on schedule. The schedule is death."

---

#### malloc_beast (Floor 5-6, /usr or /sys)

**Name:** malloc_beast - The Allocator Unchained
**Glyph:** `M`
**Stats:** PID 250 | CPU 40 | MEM 100 | NICE 8 | UID 0

**Core Mechanic:** Boss has "memory pool" of 100 points. Allocates to different abilities.

**Abilities (Cost):**
- Attack (20 memory): Standard 40 damage attack
- Fragment (10 memory): Creates obstacle that damages on contact
- Expand (30 memory): Increases own PID by 50 (max once per phase)
- Defragment (0 memory): Reclaims all fragments, heals 5 PID per fragment

**Phase 1 (100-60% PID):** 100 memory pool, slow allocation
**Phase 2 (60-30% PID):** 150 memory pool, faster allocation
**Phase 3 (30-0% PID):** 200 memory pool, "memory leak" - gains 10 memory per turn

**Arena:** Grid-based room. Fragments appear as walls/obstacles. Movement is tactical.

**Strategy:** Force defragmentation by destroying fragments. Burst damage during allocation phase.

**Drops:**
- malloc_hook (guaranteed): Once per floor, see enemy ability costs
- memory_pool (50%): +30 permanent MEM
- heap_fragment (100%, multiple): Crafting material for memory items

**Flavor:** "It allocates. It never frees. Eventually, it will allocate everything."

---

### Secret Boss

#### man_minus_one (Hidden Boss)

**Name:** man -1 - The Undocumented Command
**Glyph:** `?`
**Stats:** PID ??? | CPU ??? | MEM ??? | NICE ??? | UID ???

**Access Requirements:**
1. Find all man_page items in a single run (hidden on each floor type)
2. Use them in order: man 1, man 2, man 3... man 8
3. Then use "man -1" in /dev/null before fighting kernel_panic
4. Secret door opens to `/?` - the undocumented directory

**Arena:** A black void with floating text. The entire fight is rendered in man page format.

**Mechanic - *Undefined Documentation*:**
- Boss stats are literally unknown - no UI shows them
- Player must deduce stats through combat feedback
- Boss changes "section" periodically, altering behavior completely

**Sections (man page style):**
- Section 1 (User Commands): Standard attacks, predictable
- Section 2 (System Calls): Attacks affect player systems (debuffs)
- Section 3 (Library Functions): Buffs itself, summons helpers
- Section 5 (File Formats): Terrain manipulation, format attacks
- Section 8 (Admin Commands): Gains UID powers, permission attacks

**Phase Transitions:**
- Every 25% PID lost, section changes randomly
- At 10% PID, enters "Section ?" - all sections active simultaneously

**Special Attacks:**
- `ERRORS: This attack has undefined behavior`
- `SEE ALSO: death(1), doom(8), despair(5)`
- `BUGS: The damage cannot be avoided`

**Strategy:** Pure skill and adaptation. No reliable patterns. Must read feedback carefully.

**Drops:**
- forbidden_knowledge (guaranteed): Reveals one hidden game mechanic permanently
- man_minus_one_page (guaranteed): Lore item, explains secret ending
- undefined_weapon (50%): Weapon with randomized stats each combat, always powerful
- developer_key (10%): Unlocks debug room in main menu (cosmetic gallery, stats)

**Secret Ending:**
Defeating man_minus_one reveals that the dungeon is a corrupted man page database. The player was always documentation - a guide for future processes. Choice: become the new manual (end), or return to the system knowing the truth (continue to endless mode with permanent buff).

**Flavor:**
```
MAN(?)                   FORBIDDEN MANUAL                  MAN(?)

NAME
       man -1 - query the undocumented

SYNOPSIS
       man -1 [question...]

DESCRIPTION
       Some things are not meant to be documented.
       Some things document themselves.
       You are reading this. Therefore, it exists.
       If it exists, it can be queried.
       If it can be queried, it can be fought.
       If it can be fought, it can be defeated.
       Probably.

AUTHOR
       Unknown. Unknowable. You, possibly. Not anymore.
```

---

## Implementation Priority

### Phase 1 (Core Expansion)
1. Floor variants (high impact, moderate effort)
2. New enemies (high impact, moderate effort)
3. Mini-bosses (high impact, high effort)

### Phase 2 (Depth)
4. New items (moderate impact, low effort)
5. New character classes (high impact, high effort)
6. Side dungeons - /proc only (high impact, high effort)

### Phase 3 (Endgame)
7. Endless mode (high impact, very high effort)
8. Remaining side dungeons
9. Secret boss

### Phase 4 (Polish)
10. Balance pass on all content
11. Achievement/unlock integration
12. Lore and flavor text completion

---

*Document Version: 1.0*
*Last Updated: 2025*
*Author: /dev/dungeon Design Team*
