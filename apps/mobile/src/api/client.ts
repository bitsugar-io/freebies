// API client for Freebies backend
import * as SecureStore from 'expo-secure-store';
import { HttpError } from './errors';
import { observeServerDate } from '../utils/serverClock';

export { HttpError } from './errors';

const API_BASE_URL = __DEV__
  ? 'http://localhost:8080/api/v1'
  : 'https://freebie-api.bitsugar.io/api/v1';

const AUTH_TOKEN_KEY = 'freebie_auth_token';
const REQUEST_TIMEOUT_MS = 15_000;

export interface League {
  id: string;
  name: string;
  icon: string;
  displayOrder: number;
}

/** Pixel-art sprite served by the API: rows of palette keys, '.' = transparent. */
export interface Sprite {
  id: string;
  name: string;
  width: number;
  height: number;
  pixels: string[];
  palette: Record<string, string>;
}

export interface Event {
  id: string;
  offerId: string;
  teamId: string;
  teamName: string;
  league: string;
  teamColor?: string;
  icon?: string;
  partnerName: string;
  offerName: string;
  offerDescription: string;
  triggerCondition: string;
  regionCode?: string;
  regionName?: string;
  offerUrl?: string;
  affiliateUrl?: string;
  affiliateTagline?: string;
  isActive: boolean;
  spriteId?: string;
  sprite?: Sprite;
  /** Only fans at the game can claim it; the app tags it AT THE GAME. */
  inVenue?: boolean;
  /** Where, for an at-the-game deal ("Intuit Dome"). */
  venue?: string;
}

export interface User {
  id: string;
  deviceId: string;
  platform: string;
  pushToken?: string;
  token?: string; // Auth token, only returned on creation
  /** RFC3339 time alerts are paused until, or absent/null when not paused. */
  alertsPausedUntil?: string | null;
}


/** One LA team for the season board (GET /teams). Sorted in-season first by the server. */
export interface Team {
  id: string;
  name: string;
  fullName: string;
  league: string;
  /** Open string: a new sport needs no app release. Unknown sports get a generic badge. */
  sport: string;
  /** Display name of the sport when the API sends one ("Hockey"). Otherwise derived from `sport`. */
  sportName?: string;
  /** Display league, "NHL". */
  leagueName?: string;
  badgeSpriteId?: string;
  /** Two-tone team identity (NES hex) for team tiles only. Optional. */
  colors?: { primary: string; secondary: string };
  /** Badge art from the API (same shape as event.sprite). Bundled badges are offline fallback. */
  badge?: Sprite;
  inSeason: boolean;
  /** When an off-season team returns: "MAR", "FEB 11". Null in season. */
  returnLabel: string | null;
  activeDeals: number;
}

/** A deal anyone can claim right now (any LA team), from GET /today. */
export interface TodayDeal {
  id: string;
  eventId: string;
  triggeredAt: string;
  availableAt?: string;
  expiresAt?: string;
  event: Event;
}

/** One of today's LA games with deals on the line. */
export interface TodayGame {
  teamId: string;
  teamName: string;
  league: string;
  opponent: string;
  startsAt: string;
  home: boolean;
  state: 'pre' | 'in' | 'post';
  deals: Event[];
}

export interface Today {
  /** Pacific date, YYYY-MM-DD. */
  date: string;
  live: TodayDeal[];
  games: TodayGame[];
}

export interface Subscription {
  id: string;
  userId: string;
  eventId: string;
  event: Event;
}

export interface ActiveDeal {
  id: string;
  eventId: string;
  triggeredAt: string;
  /** RFC3339 time the code unlocks, when it unlocks after the game (Rams: 2 days later). */
  availableAt?: string;
  expiresAt?: string;
  event: Event;
  isDismissed: boolean;
  dismissalType?: 'got_it' | 'stop_reminding';
}

/** A deal the user claimed. It stays listed after its window closes (the trophy case). */
export interface ClaimedDeal {
  /** Triggered event id. */
  id: string;
  eventId: string;
  /** RFC3339 time the user tapped GOT IT. */
  claimedAt: string;
  triggeredAt: string;
  /** RFC3339 time the code unlocked, when it unlocks after the game. */
  availableAt?: string;
  event: Event;
}

export interface Dismissal {
  id: string;
  userId: string;
  triggeredEventId: string;
  type: 'got_it' | 'stop_reminding';
  dismissedAt: string;
}

export interface ScreenBlock {
  type: string;
  key: string;
  config: Record<string, any>;
}

export interface AppConfig {
  features: Record<string, boolean>;
  screens: Record<string, ScreenBlock[]>;
}

export interface ApiError {
  error: string;
}

class ApiClient {
  private baseUrl: string;
  private authToken: string | null = null;

  constructor(baseUrl: string = API_BASE_URL) {
    this.baseUrl = baseUrl;
  }

  // Load auth token from secure storage
  async loadAuthToken(): Promise<string | null> {
    this.authToken = await SecureStore.getItemAsync(AUTH_TOKEN_KEY);
    return this.authToken;
  }

  // Save auth token to secure storage
  async setAuthToken(token: string): Promise<void> {
    this.authToken = token;
    await SecureStore.setItemAsync(AUTH_TOKEN_KEY, token);
  }

  // Clear auth token
  async clearAuthToken(): Promise<void> {
    this.authToken = null;
    await SecureStore.deleteItemAsync(AUTH_TOKEN_KEY);
  }

  // Get current auth token
  getAuthToken(): string | null {
    return this.authToken;
  }

  /** The API root, e.g. https://freebie-api.bitsugar.io/api/v1 (shared with the widget). */
  getBaseUrl(): string {
    return this.baseUrl;
  }

  private async request<T>(
    path: string,
    options: RequestInit = {},
    authenticated: boolean = false
  ): Promise<T> {
    const url = `${this.baseUrl}${path}`;
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      ...(options.headers as Record<string, string>),
    };

    // Add auth header for authenticated requests
    if (authenticated && this.authToken) {
      headers['Authorization'] = `Bearer ${this.authToken}`;
    }

    // Time out instead of hanging forever on a bad connection.
    const controller = new AbortController();
    const timeout = setTimeout(() => controller.abort(), REQUEST_TIMEOUT_MS);
    let response: Response;
    const sentAt = Date.now();
    try {
      response = await fetch(url, { ...options, headers, signal: controller.signal });
      // Correct the device clock from the server's (expiry and countdowns use it).
      observeServerDate(response.headers?.get?.('Date'), sentAt, Date.now());
    } catch (err) {
      if (controller.signal.aborted) throw new Error('Request timed out');
      throw err;
    } finally {
      clearTimeout(timeout);
    }

    if (!response.ok) {
      const error: ApiError = await response.json().catch(() => ({
        error: `HTTP ${response.status}`,
      }));
      throw new HttpError(response.status, error.error || `Request failed: ${response.status}`);
    }

    // Handle 204 No Content
    if (response.status === 204) {
      return undefined as T;
    }

    return response.json();
  }

  // Leagues
  async listLeagues(): Promise<League[]> {
    return this.request<League[]>('/leagues');
  }

  // Events
  async listEvents(): Promise<Event[]> {
    return this.request<Event[]>('/events');
  }

  async getEvent(id: string): Promise<Event> {
    return this.request<Event>(`/events/${id}`);
  }

  // Today: claimable deals from every LA team and today's games
  async getToday(): Promise<Today> {
    return this.request<Today>('/today');
  }

  /** One triggered deal by id (public; same shape as a TODAY live item). Expired ones too. */
  async getTriggered(id: string): Promise<TodayDeal> {
    return this.request<TodayDeal>(`/triggered/${encodeURIComponent(id)}`);
  }

  // Teams
  async listTeams(): Promise<Team[]> {
    return this.request<Team[]>('/teams');
  }

  // Users
  async createUser(deviceId: string, platform: string): Promise<User> {
    const user = await this.request<User>('/users', {
      method: 'POST',
      body: JSON.stringify({ deviceId, platform }),
    });
    // Store token for future authenticated requests
    if (user.token) {
      await this.setAuthToken(user.token);
    }
    return user;
  }

  async getUser(id: string): Promise<User> {
    return this.request<User>(`/users/${id}`, {}, true);
  }

  async updatePushToken(userId: string, pushToken: string): Promise<void> {
    return this.request<void>(
      `/users/${userId}/push-token`,
      {
        method: 'PUT',
        body: JSON.stringify({ pushToken }),
      },
      true
    );
  }

  /** Pause alerts until a time (RFC3339), or resume with null. Returns the stored value. */
  async setAlertsPause(userId: string, until: string | null): Promise<{ pausedUntil: string | null }> {
    return this.request<{ pausedUntil: string | null }>(
      `/users/${userId}/alerts-pause`,
      { method: 'PUT', body: JSON.stringify({ until }) },
      true
    );
  }

  // Subscriptions
  async listSubscriptions(userId: string): Promise<Subscription[]> {
    return this.request<Subscription[]>(`/users/${userId}/subscriptions`, {}, true);
  }

  async createSubscription(
    userId: string,
    eventId: string
  ): Promise<Subscription> {
    return this.request<Subscription>(
      `/users/${userId}/subscriptions`,
      {
        method: 'POST',
        body: JSON.stringify({ eventId }),
      },
      true
    );
  }

  async deleteSubscription(userId: string, eventId: string): Promise<void> {
    return this.request<void>(
      `/users/${userId}/subscriptions/${eventId}`,
      { method: 'DELETE' },
      true
    );
  }

  // Active Deals
  async listActiveDeals(userId: string): Promise<ActiveDeal[]> {
    return this.request<ActiveDeal[]>(`/users/${userId}/active-deals`, {}, true);
  }

  async listClaimed(userId: string, limit?: number): Promise<ClaimedDeal[]> {
    const query = limit ? `?limit=${limit}` : '';
    return this.request<ClaimedDeal[]>(`/users/${userId}/claimed${query}`, {}, true);
  }

  // Dismissals
  async createDismissal(
    userId: string,
    triggeredEventId: string,
    type: 'got_it' | 'stop_reminding' = 'got_it'
  ): Promise<Dismissal> {
    return this.request<Dismissal>(
      `/users/${userId}/dismissals`,
      {
        method: 'POST',
        body: JSON.stringify({ triggeredEventId, type }),
      },
      true
    );
  }

  async deleteDismissal(userId: string, triggeredEventId: string): Promise<void> {
    return this.request<void>(
      `/users/${userId}/dismissals/${triggeredEventId}`,
      { method: 'DELETE' },
      true
    );
  }

  // User Stats
  async getUserStats(userId: string): Promise<UserStats> {
    return this.request<UserStats>(`/users/${userId}/stats`, {}, true);
  }

  // App Config
  async getConfig(): Promise<AppConfig> {
    return this.request<AppConfig>('/config');
  }
}

export interface UserStats {
  dealsClaimed: number;
  subscriptionsCount: number;
}

export const api = new ApiClient();
