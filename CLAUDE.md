# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

# Target Auto-Checkout Bot - Claude Code Context

## 1. Project Overview & Goal
Build a high-performance, concurrent automated checkout bot for Target. The system must bypass PerimeterX (HUMAN Security) using TLS spoofing and execute rapid, API-based checkout flows. This is a monolithic, backend-focused application built for speed and reliability. No microservices.

## 2. Technology Stack
- **Language:** Go (1.21+)
- **Networking/Anti-Bot:** MUST use `github.com/bogdanfinn/tls-client` (or similar `utls` wrapper) to spoof JA3/HTTP2 fingerprints and bypass PerimeterX. Do not use the standard `net/http` client for Target API calls.
- **Concurrency:** Native Goroutines and Channels.
- **Data Parsing:** Standard `encoding/json`.

## 3. Architecture & OOP Design
Follow clean architecture principles using Go idiomatic pseudo-OOP (Structs and Interfaces).
- **`cmd/`**: Entry point (CLI). Parses user inputs and task CSVs.
- **`internal/orchestrator/`**: Manages the application lifecycle, initializes the Monitor, and spins up the Worker Pool.
- **`internal/monitor/`**: Responsible strictly for polling Target's DPCI inventory API. Uses high-quality proxies.
- **`internal/session/`**: Manages the TLS client, proxy rotation, and stores `_px3` cookies. Every task gets its own isolated session.
- **`internal/task/`**: The state machine for a single checkout attempt. 

## 4. Design Patterns & Concurrency
- **Worker Pool Pattern:** Do not spawn unlimited goroutines. The Orchestrator should create a fixed pool of `CheckoutWorker` goroutines.
- **Observer Pattern (via Channels):** The `Monitor` runs in the background. When it detects stock, it broadcasts the payload down a `StockEvent` channel. Idle workers listen on this channel and instantly begin the checkout flow.
- **State Machine:** Use Go `iota` constants to define strict task states: `StateIdle`, `StateAddingToCart`, `StateSubmittingPayment`, `StateSuccess`, `StateFailed`.

## 5. Coding Style & Standards
- **Idiomatic Go:** Adhere to "Effective Go". Use short variable names for narrow scopes, descriptive names for package-level variables.
- **Error Handling:** Explicit error checking (`if err != nil`). Wrap errors with context (`fmt.Errorf("failed to add to cart: %w", err)`). NEVER use `panic` outside of initialization.
- **Interfaces:** Define behavior via interfaces (e.g., `type CheckoutClient interface { AddToCart(...) error }`) to allow for easy mock testing.
- **Documentation:** Provide high-level comments for all exported structs and methods.
