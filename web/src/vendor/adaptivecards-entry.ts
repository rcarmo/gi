/**
 * vendor/adaptivecards-entry.ts – Browser entry point for vendored Adaptive Cards SDK.
 *
 * Exposes the AdaptiveCard class, HostConfig, and related types as globals
 * for lazy-loading by the timeline renderer.
 */
import * as AdaptiveCards from "adaptivecards";

(globalThis as any).AdaptiveCards = AdaptiveCards;
