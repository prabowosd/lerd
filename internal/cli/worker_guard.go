package cli

import "fmt"

// buildWorkerGuard wraps runCmd in a shell snippet that prevents duplicate
// workers on macOS under the podman-exec runtime mode.
//
// Two failure modes are addressed:
//
//  1. Brief podman-machine SSH hiccup. The outer `podman exec` exits but
//     the inner artisan process inside the FPM container survives. The
//     pid-file mutex (step 1) catches the case where launchd respawns
//     before the outer process is gone.
//
//  2. Suspend/wake. The laptop sleeps; on wake the host-side `podman exec`
//     dies (its TCP/vsock link to the machine was torn down) but the inner
//     artisan process inside the container resumed normally. The pid-file
//     mutex doesn't help — its EXIT trap removed the file when the outer
//     process died — so step 2 reaches into the container with `pkill -f`
//     to graceful-stop any orphan worker matching the command line, then
//     step 3 launches a fresh one. Laravel/Symfony workers respond to
//     SIGTERM by finishing the current job before exiting, so the queue
//     layer never sees a half-processed message.
//
// On launch:
//
//  1. If the pid file exists AND its PID is alive, the previous outer
//     process is still driving the worker — exit 0.
//  2. Otherwise pkill any container-side process matching workerCmd
//     (the inner Laravel/Symfony command, sans `podman exec` prefix).
//     Failures are swallowed — pkill returns non-zero when nothing
//     matches, which is the common case.
//  3. Record our own PID, install an EXIT trap to clean up, and replace
//     ourselves with runCmd so signals (TERM from launchd on shutdown)
//     reach it directly.
//
// Stale pid files (previous process crashed) resolve on their own: the
// kill -0 check in step 1 fails and the new instance takes over.
func buildWorkerGuard(pidFile, podmanBin, container, workerCmd, runCmd string) string {
	return fmt.Sprintf(`if [ -f %[1]s ] && kill -0 "$(cat %[1]s 2>/dev/null)" 2>/dev/null; then
  exit 0
fi
%[2]s exec %[3]s pkill -f -- %[4]s >/dev/null 2>&1 || true
echo $$ > %[1]s
trap 'rm -f %[1]s' EXIT
exec %[5]s
`, pidFile, podmanBin, container, shellQuote(workerCmd), runCmd)
}

// shellQuote single-quotes s for safe inclusion as one shell argument.
// Embedded single quotes are handled by closing-quoting-reopening.
func shellQuote(s string) string {
	out := "'"
	for _, r := range s {
		if r == '\'' {
			out += `'\''`
			continue
		}
		out += string(r)
	}
	out += "'"
	return out
}
