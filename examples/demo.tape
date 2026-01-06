# MAGE-X Demo

Output demo.gif

Require magex
Require mgx

# ═══════════════════════════════════════════════════════════════════
# VISUAL SETTINGS
# ═══════════════════════════════════════════════════════════════════

Set Shell "zsh"
Set FontSize 15
Set Width 1100
Set Height 850
Set Padding 20
Set BorderRadius 10
Set WindowBar Colorful
Set TypingSpeed 50ms
Set LineHeight 1.2

# ═══════════════════════════════════════════════════════════════════
# MAGE-X THEME
# ═══════════════════════════════════════════════════════════════════

Set Theme {
  "name": "CyberMage",
  "black": "#0d0d0d",
  "red": "#ff5f87",
  "green": "#50fa7b",
  "yellow": "#f1fa8c",
  "blue": "#6be5fd",
  "magenta": "#ff79c6",
  "cyan": "#8be9fd",
  "white": "#f8f8f2",
  "brightBlack": "#6272a4",
  "brightRed": "#ff6e9c",
  "brightGreen": "#69ff94",
  "brightYellow": "#ffffa5",
  "brightBlue": "#a4ffff",
  "brightMagenta": "#ff92df",
  "brightCyan": "#a4ffff",
  "brightWhite": "#ffffff",
  "background": "#0d0d0d",
  "foreground": "#50fa7b",
  "selection": "#ff79c655",
  "cursor": "#ff79c6"
}

# ═══════════════════════════════════════════════════════════════════
# EXAMPLE / DEMO SCRIPT
# ═══════════════════════════════════════════════════════════════════

# Scene 1: Version check
Sleep 500ms
Type "magex --version"
Sleep 300ms
Enter
Sleep 1.5s

Hide
Type "clear"
Enter
Show

# Scene 2: Help overview
Type "mgx help"
Sleep 300ms
Enter
Sleep 3s

Hide
Type "clear"
Enter
Show

# Scene 3: Format fix
Type "mgx format:fix"
Sleep 300ms
Enter
Sleep 5s

Hide
Type "clear"
Enter
Show

# Scene 4: Lint
Type "mgx lint"
Sleep 300ms
Enter
Sleep 5s

Hide
Type "clear"
Enter
Show

# Scene 5: Test
Type "mgx test"
Sleep 300ms
Enter
Sleep 5s

Hide
Type "clear"
Enter
Show

# Scene 6: Metrics
Type "mgx metrics:loc"
Sleep 300ms
Enter
Sleep 5s
