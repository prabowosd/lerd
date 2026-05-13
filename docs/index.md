---
layout: home
title: Lerd - Local PHP development for Linux
description: Open-source local PHP development environment built for Linux. Nginx + PHP-FPM 8.1–8.5 on rootless Podman, with automatic .test domains, HTTPS, and a built-in Web UI. First-class on Arch, Ubuntu, Fedora, and Debian; also runs on macOS via Homebrew.

hero:
  name: Lerd
  text: Local PHP development for Linux
  tagline: Nginx + PHP-FPM 8.1–8.5 on rootless Podman. Drop any project in for automatic .test domains and HTTPS, no config files, no Docker daemon, no sudo. First-class on Arch, Ubuntu, Fedora, and Debian; macOS supported too.
  image:
    src: /assets/screenshots/app-1.png
    alt: Lerd dashboard
  actions:
    - theme: brand
      text: Get Started
      link: /getting-started/requirements
    - theme: alt
      text: View on GitHub
      link: https://github.com/geodro/lerd

features:
  - icon: 🤖
    title: AI integration (MCP)
    details: Built-in Model Context Protocol server. Claude Code, Cursor, Windsurf, and Junie can scaffold projects, run migrations, and tail logs straight from chat.
  - icon: 📦
    title: Rootless Podman
    details: No Docker daemon, no sudo for containers, no system pollution. Services run as your user via rootless Podman and systemd user units on Linux and macOS.
  - icon: 🌐
    title: Auto .test domains
    details: Every linked project gets a .test domain instantly, with dual-stack IPv4 and IPv6 and no /etc/hosts edits or DNS setup required on your machine.
  - icon: 🐘
    title: PHP & Node versions
    details: PHP 8.1–8.5 plus a frozen 7.4 / 8.0 legacy tier, with multiple Node versions side by side and per-project pinning from CLI, UI, or TUI.
  - icon: 🔒
    title: One-command HTTPS
    details: "`lerd secure` issues a locally-trusted TLS cert via mkcert, rewrites the nginx vhost, and updates APP_URL for you in a single shot."
  - icon: ⚡
    title: FrankenPHP runtime
    details: Per-site FrankenPHP as an alternative to shared PHP-FPM, with Laravel Octane and Symfony Runtime worker mode wired up out of the box.
  - icon: 🔧
    title: Services & presets
    details: Built-ins (MySQL, Postgres, Redis, Meilisearch, RustFS, Mailpit), one-click presets, or any OCI image as a custom service with depends_on cascades.
  - icon: 🖥️
    title: Web UI, TUI & tray
    details: Browser dashboard with command palette, btop-style TUI (`lerd tui`), and a system tray applet, all wired to the same live event bus.
  - icon: 🧪
    title: Tinker tab
    details: In-browser PHP REPL per site with autocomplete, live linting, and collapsible dump trees. Works on Laravel, Symfony, and any composer-based project.
  - icon: 🛰️
    title: Live dump() / dd() viewer
    details: Every `dump()` and `dd()` call from your running app streams to the dashboard, TUI, MCP, and CLI tail, scoped per site and per worktree branch.
  - icon: 🧩
    title: Framework store
    details: YAML framework definitions for Laravel, Symfony, WordPress, Drupal, CakePHP, and Statamic, auto-detected and applied on `lerd link`.
  - icon: 🧱
    title: Polyglot sites
    details: Drop a `Containerfile.lerd` to run Node, Python, Ruby, or Go sites alongside your PHP ones, with the same HTTPS, DNS, and worker pipeline.
---
