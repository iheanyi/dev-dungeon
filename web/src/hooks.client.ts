import * as Sentry from '@sentry/sveltekit';
import { env } from '$env/dynamic/public';

const dsn = env.PUBLIC_SENTRY_DSN;

if (dsn) {
	Sentry.init({
		dsn,
		environment: env.PUBLIC_SENTRY_ENVIRONMENT || 'development',

		// Sample 10% of transactions for performance monitoring
		tracesSampleRate: 0.1,

		// Capture Replay for 10% of sessions, 100% of sessions with errors
		replaysSessionSampleRate: 0.1,
		replaysOnErrorSampleRate: 1.0,

		integrations: [Sentry.replayIntegration()]
	});
}

export const handleError = Sentry.handleErrorWithSentry();
