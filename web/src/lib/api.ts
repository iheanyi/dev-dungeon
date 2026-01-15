// API client for /dev/dungeon web portal

const API_BASE = '/api';

export interface ApiResponse<T> {
	success: boolean;
	data?: T;
	error?: string;
}

export interface LeaderboardEntry {
	username: string;
	run_type: string;
	score: number;
	floors_cleared: number;
	time_seconds: number;
	class: string;
	created_at: string;
	rank?: number;
}

export interface PlayerProfile {
	username: string;
	public_id: string;
	created_at: string;
	runs_completed: number;
	deepest_floor: number;
	total_deaths: number;
	unlocked_classes?: string[];
}

export interface DailySeed {
	date: string;
	seed: number;
}

export interface User {
	username: string;
	public_id: string;
	created_at: string;
	runs_completed?: number;
	deepest_floor?: number;
	total_deaths?: number;
	unlocked_classes?: string[];
	total_exit_codes?: number;
}

export interface RegisterRequest {
	username: string;
	public_key: string;
}

export interface RegisterResponse {
	username: string;
	public_id: string;
	message: string;
}

async function fetchApi<T>(endpoint: string, options?: RequestInit): Promise<ApiResponse<T>> {
	try {
		const response = await fetch(`${API_BASE}${endpoint}`, {
			headers: {
				'Content-Type': 'application/json',
			},
			credentials: 'include', // Include cookies for auth
			...options,
		});

		const data = await response.json();
		return data;
	} catch {
		return {
			success: false,
			error: 'Network error. Please try again.',
		};
	}
}

export async function getLeaderboard(runType?: string): Promise<ApiResponse<LeaderboardEntry[]>> {
	const endpoint = runType ? `/leaderboard/${runType}` : '/leaderboard';
	return fetchApi<LeaderboardEntry[]>(endpoint);
}

export async function getPlayerProfile(username: string): Promise<ApiResponse<PlayerProfile>> {
	return fetchApi<PlayerProfile>(`/players/${username}`);
}

export async function getDailySeed(): Promise<ApiResponse<DailySeed>> {
	return fetchApi<DailySeed>('/daily');
}

export async function register(username: string, publicKey: string): Promise<ApiResponse<RegisterResponse>> {
	return fetchApi<RegisterResponse>('/register', {
		method: 'POST',
		body: JSON.stringify({
			username,
			public_key: publicKey,
		}),
	});
}

export function formatTime(seconds: number): string {
	const mins = Math.floor(seconds / 60);
	const secs = seconds % 60;
	return `${mins}:${secs.toString().padStart(2, '0')}`;
}

export function formatDate(dateString: string): string {
	return new Date(dateString).toLocaleDateString('en-US', {
		year: 'numeric',
		month: 'short',
		day: 'numeric',
	});
}

export const FLOOR_NAMES = [
	'/home',
	'/tmp',
	'/var',
	'/etc',
	'/usr',
	'/sys',
	'/dev',
	'/dev/null',
];

export function getFloorName(floor: number): string {
	return FLOOR_NAMES[floor] || `Floor ${floor}`;
}

// Auth functions
export async function getCurrentUser(): Promise<ApiResponse<User>> {
	return fetchApi<User>('/auth/me');
}

export async function logout(): Promise<ApiResponse<{ message: string }>> {
	return fetchApi<{ message: string }>('/auth/logout', {
		method: 'POST',
	});
}
