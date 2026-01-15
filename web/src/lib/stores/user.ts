// User auth state store
import { writable } from 'svelte/store';
import { logout as apiLogout } from '$lib/api';
import { browser } from '$app/environment';

interface UserState {
	username: string | null;
	loading: boolean;
	checked: boolean;
}

// Read username from cookie (non-HttpOnly, set by server)
function getUsernameFromCookie(): string | null {
	if (!browser) return null;

	const cookies = document.cookie.split(';');
	for (const cookie of cookies) {
		const [name, value] = cookie.trim().split('=');
		if (name === 'devdungeon_user') {
			return decodeURIComponent(value);
		}
	}
	return null;
}

function createUserStore() {
	const { subscribe, set } = writable<UserState>({
		username: null,
		loading: true,
		checked: false,
	});

	return {
		subscribe,

		// Check auth from cookie (instant, no API call)
		checkAuth() {
			const username = getUsernameFromCookie();
			set({ username, loading: false, checked: true });
		},

		// Log out - clears cookie via API
		async logout() {
			await apiLogout();
			set({ username: null, loading: false, checked: true });
		},

		// Clear state
		clear() {
			set({ username: null, loading: false, checked: true });
		}
	};
}

export const userStore = createUserStore();
