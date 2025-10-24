Output gif/000save.gif

Source "_config.tape"
Source "_set_envs.tape"

Sleep 500ms
Type 'echo "You can specify a key." | wow howto/save'
Sleep 500ms
Enter
Sleep 2s

Type 'echo "Or let one generate." | wow'
Sleep 500ms
Enter
Sleep 2s

Type 'clear'
Sleep 500ms
Enter
Sleep 500ms

Type 'echo "Any sort of stdin works." > foo.txt'
Sleep 500ms
Enter
Sleep 1s
Type 'wow foo < foo.txt'
Sleep 500ms
Enter
Sleep 1s

Type 'clear'
Sleep 500ms
Enter
Sleep 500ms

Type 'echo "Saving prints the key." | wow foo_bar'
Sleep 500ms
Enter
Sleep 1s
Type 'But sometimes you want to --tee the content elsewhere.'
Backspace 58
Type 'echo "So use the tee flag" | wow baz --tee'
Sleep 500ms
Enter
Sleep 2s
