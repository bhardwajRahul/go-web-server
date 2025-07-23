// Package ui provides embedded static assets for the web application.
package ui

import "embed"

// StaticFiles contains all embedded static assets (CSS, JS, images, etc.).
//
//go:embed static/*
var StaticFiles embed.FS
