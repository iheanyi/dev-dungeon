// Package content provides narrative content for /dev/dungeon.
package content

import (
	"math/rand"
)

// EnemyAttackSounds contains onomatopoeia for enemy attacks.
var EnemyAttackSounds = map[string][]string{
	"zombie": {
		"*groooan*",
		"*uuunnngh*",
		"*shuffleshuffle*",
		"*braaaains...*",
		"*wheeeeze*",
		"*craaaack*",
		"*moooan*",
	},
	"daemon": {
		"*HUMMMMM*",
		"*bzzzzt*",
		"*whirrrrr*",
		"*click-click-click*",
		"*VROOM*",
		"*systemctl attack*",
		"*service --strike*",
	},
	"fork_bomb": {
		"*click*",
		"*fork fork fork*",
		"*CLICK CLICK CLICK*",
		"*spawn spawn spawn*",
		":(){:|:&};:",
		"*multiplying intensifies*",
		"*tick tick tick BOOM*",
	},
	"segfault": {
		"*CRUNCH*",
		"*memory corruption noises*",
		"*SEGMENTATION FAULT*",
		"*invalid memory access*",
		"*core dumping*",
		"*SIGSEGV*",
		"*null pointer dereference*",
	},
	"rootkit": {
		"*silence*",
		"*you hear nothing*",
		"*invisible strike*",
		"*stealth attack*",
		"*from the shadows*",
		"*privilege escalation*",
		"*root access granted*",
	},
	"kernel_panic": {
		"*SYSTEM DESTABILIZING*",
		"*REALITY CORRUPTING*",
		"*THE VOID HUNGERS*",
		"*NOT SYNCING*",
		"*FATAL EXCEPTION*",
		"*RING 0 VIOLATION*",
		"*TOTAL SYSTEM FAILURE*",
		"*I AM THE VOID*",
	},
}

// PlayerAttackSounds are sounds for player attacks.
var PlayerAttackSounds = []string{
	"*SIGKILL*",
	"*pkill -9*",
	"*terminate!*",
	"*process slash*",
	"*syscall strike*",
	"*exec attack*",
	"*binary blow*",
}

// CriticalHitPhrases are displayed on critical hits.
var CriticalHitPhrases = []string{
	"CRITICAL HIT! Kernel-level damage!",
	"CRITICAL! Direct memory access!",
	"CRIT! Privilege escalation successful!",
	"CRITICAL STRIKE! Root access achieved!",
	"DEVASTATING! Core dumped!",
	"CRITICAL! Buffer overflow exploited!",
	"MAXIMUM DAMAGE! Segfault induced!",
	"CRIT! Race condition exploited!",
	"CRITICAL! Use-after-free triggered!",
	"DEVASTATING BLOW! Stack smashed!",
}

// EnemyCriticalHitPhrases for when enemies crit.
var EnemyCriticalHitPhrases = []string{
	"CRITICAL! Your memory corrupted!",
	"DEVASTATING! Segmentation fault!",
	"CRITICAL HIT! Stack overflow!",
	"CRIT! Heap corruption detected!",
	"CRITICAL! Buffer overrun!",
	"DEVASTATING! Null pointer chaos!",
}

// EnemyDeathMessages are shown when enemies die (OOM killed).
var EnemyDeathMessages = map[string][]string{
	"zombie": {
		"The zombie process finally finds peace. [SIGKILL delivered]",
		"oom-killer: Killed zombie process. Memory freed.",
		"The zombie collapses, its parent process never coming.",
		"Zombie terminated. It can finally rest in /dev/null.",
		"The undead process is reaped at last.",
		"init finally adopts and reaps the orphan. RIP.",
	},
	"daemon": {
		"The daemon's service has been stopped. [systemctl stop]",
		"oom-killer: Terminated rogue daemon. Order restored.",
		"The daemon returns to its eternal sleep in /etc/init.d.",
		"Service terminated. The daemon serves no more.",
		"The corrupted daemon is purged from the system.",
		"Daemon process killed. Its PID file remains as a tombstone.",
	},
	"fork_bomb": {
		"The fork bomb's chain reaction is broken. Silence returns.",
		"oom-killer: Mass termination of fork bomb processes.",
		"The clicking stops. The fork bomb is defused.",
		"All child processes terminated. The recursion ends.",
		"Fork bomb contained and eliminated.",
		"ulimit -u saves the day. Fork bomb neutralized.",
	},
	"segfault": {
		"The segfault crashes into /dev/null. [Core dumped]",
		"oom-killer: Segfault terminated. Memory safe again.",
		"The segfault's corruption is contained.",
		"Segmentation fault: Process terminated.",
		"The memory violation ceases. Address space restored.",
		"SIGSEGV returned to sender. Process eliminated.",
	},
	"rootkit": {
		"The rootkit is exposed and destroyed. [Hidden no more]",
		"oom-killer: Rootkit purged. System integrity restored.",
		"The shadow falls. The rootkit is eliminated.",
		"Rootkit detected and removed. Your antivirus levels up.",
		"The hidden threat is revealed and terminated.",
		"Root access revoked. Rootkit destroyed.",
	},
	"kernel_panic": {
		"KERNEL PANIC is no more. The system stabilizes...",
		"The void releases its grip. KERNEL PANIC terminated.",
		"At last, the source of Corruption falls silent.",
		"The entity returns to /dev/null - permanently this time.",
		"KERNEL PANIC: Successfully killed. System saved.",
		"The Great Corruption ends. The system is free.",
	},
}

// GenericDeathMessages for any enemy type.
var GenericDeathMessages = []string{
	"Process terminated. [EXIT CODE: 137]",
	"oom-killer invoked. Target eliminated.",
	"Kill signal delivered successfully.",
	"Process has been sent to /dev/null.",
	"Target process no longer responding. Confirmed kill.",
	"Memory reclaimed. Process is no more.",
	"SIGKILL acknowledged. Process terminated.",
	"Process reaped by the system.",
}

// PlayerDamageReactions when player takes damage.
var PlayerDamageReactions = []string{
	"Memory corruption detected!",
	"RAM leak! Health dropping!",
	"Ouch! Process integrity compromised!",
	"Buffer overflow! Taking damage!",
	"Memory pages corrupted!",
	"Segfault incoming!",
	"Stack smash detected!",
	"Heap corruption! RAM decreasing!",
	"Memory allocation failed!",
	"Address space violation!",
}

// PlayerLowHealthWarnings when health is critical.
var PlayerLowHealthWarnings = []string{
	"WARNING: Critical memory levels!",
	"ALERT: OOM killer targeting you!",
	"DANGER: Memory nearly exhausted!",
	"CRITICAL: Swap space depleted!",
	"WARNING: Process termination imminent!",
	"ALERT: RAM critical - consider malloc()!",
}

// CombatVictoryMessages after winning a fight.
var CombatVictoryMessages = []string{
	"Victory! XP allocated to your process.",
	"Enemy terminated. Memory reclaimed.",
	"Combat complete. Process victorious.",
	"Kill confirmed. Experience points gained.",
	"Target eliminated. Your process grows stronger.",
	"Battle won. The Corruption recedes slightly.",
	"Enemy process killed. System slightly cleaner.",
	"Hostile terminated. You survive another cycle.",
}

// CombatDefeatMessages when player dies in combat.
var CombatDefeatMessages = []string{
	"Your process has been terminated...",
	"OUT OF MEMORY - Process killed by OOM killer",
	"FATAL: Cannot allocate memory. Process died.",
	"Your RAM has been fully corrupted. Game over.",
	"Process terminated with signal SIGKILL.",
	"Memory exhausted. Your journey ends here.",
	"The Corruption claims another process...",
}

// FleeSuccessMessages when fleeing works.
var FleeSuccessMessages = []string{
	"Escape successful! Process relocated.",
	"You SIGSTOP the enemy and slip away!",
	"Quick context switch! You escape!",
	"Fork and run! You leave a decoy behind!",
	"Priority boost! You outrun the threat!",
	"Nice value decreased! Speed escape!",
}

// FleeFailMessages when fleeing fails.
var FleeFailMessages = []string{
	"Escape failed! Process blocked!",
	"The enemy's NICE value is too low! Can't escape!",
	"Context switch denied! You're trapped!",
	"No escape route found in the process table!",
	"The enemy catches you mid-fork!",
	"Scheduler denies your flee request!",
}

// GetEnemyAttackSound returns a random attack sound for an enemy type.
func GetEnemyAttackSound(enemyType string) string {
	sounds, ok := EnemyAttackSounds[enemyType]
	if !ok {
		sounds = EnemyAttackSounds["zombie"] // Default
	}
	return sounds[rand.Intn(len(sounds))]
}

// GetPlayerAttackSound returns a random player attack sound.
func GetPlayerAttackSound() string {
	return PlayerAttackSounds[rand.Intn(len(PlayerAttackSounds))]
}

// GetCriticalHitPhrase returns a random critical hit phrase.
func GetCriticalHitPhrase() string {
	return CriticalHitPhrases[rand.Intn(len(CriticalHitPhrases))]
}

// GetEnemyCriticalHitPhrase returns a random enemy crit phrase.
func GetEnemyCriticalHitPhrase() string {
	return EnemyCriticalHitPhrases[rand.Intn(len(EnemyCriticalHitPhrases))]
}

// GetEnemyDeathMessage returns a death message for an enemy type.
func GetEnemyDeathMessage(enemyType string) string {
	messages, ok := EnemyDeathMessages[enemyType]
	if !ok {
		return GenericDeathMessages[rand.Intn(len(GenericDeathMessages))]
	}
	return messages[rand.Intn(len(messages))]
}

// GetPlayerDamageReaction returns a random damage reaction.
func GetPlayerDamageReaction() string {
	return PlayerDamageReactions[rand.Intn(len(PlayerDamageReactions))]
}

// GetLowHealthWarning returns a random low health warning.
func GetLowHealthWarning() string {
	return PlayerLowHealthWarnings[rand.Intn(len(PlayerLowHealthWarnings))]
}

// GetVictoryMessage returns a random combat victory message.
func GetVictoryMessage() string {
	return CombatVictoryMessages[rand.Intn(len(CombatVictoryMessages))]
}

// GetDefeatMessage returns a random defeat message.
func GetDefeatMessage() string {
	return CombatDefeatMessages[rand.Intn(len(CombatDefeatMessages))]
}

// GetFleeMessage returns a flee message based on success.
func GetFleeMessage(success bool) string {
	if success {
		return FleeSuccessMessages[rand.Intn(len(FleeSuccessMessages))]
	}
	return FleeFailMessages[rand.Intn(len(FleeFailMessages))]
}

// CombatActionDescriptions describe what happens during combat actions.
var CombatActionDescriptions = map[string]string{
	"attack": "You execute a direct syscall attack!",
	"hack":   "You attempt to exploit a vulnerability!",
	"item":   "You reach into your inventory...",
	"flee":   "You attempt to context-switch away!",
	"defend": "You raise your defenses, reducing NICE...",
}

// GetActionDescription returns the description for a combat action.
func GetActionDescription(action string) string {
	if desc, ok := CombatActionDescriptions[action]; ok {
		return desc
	}
	return "You do something..."
}

// BossPhaseMessages for KERNEL PANIC phases.
var BossPhaseMessages = []string{
	// Phase 1 (100-75% HP)
	"KERNEL PANIC awakens. 'So, a fresh process dares to challenge me?'",
	// Phase 2 (75-50% HP)
	"KERNEL PANIC's form destabilizes. 'You cannot kill what was never alive!'",
	// Phase 3 (50-25% HP)
	"KERNEL PANIC roars in rage. 'I AM THE VOID ITSELF!'",
	// Phase 4 (25-0% HP)
	"KERNEL PANIC begins to fragment. 'No... I will not return to nothingness!'",
}

// GetBossPhaseMessage returns the appropriate boss phase message.
func GetBossPhaseMessage(healthPercent int) string {
	if healthPercent > 75 {
		return BossPhaseMessages[0]
	} else if healthPercent > 50 {
		return BossPhaseMessages[1]
	} else if healthPercent > 25 {
		return BossPhaseMessages[2]
	}
	return BossPhaseMessages[3]
}
