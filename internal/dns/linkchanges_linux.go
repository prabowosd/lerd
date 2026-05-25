package dns

import (
	"errors"

	"golang.org/x/sys/unix"
)

// rtnetlink legacy multicast group masks. Using the pre-shifted RTMGRP_*
// constants (as opposed to position-based RTNLGRP_*) lets us OR them
// straight into SockaddrNetlink.Groups, which is the legacy 32-bit mask
// field, without per-group bit shifting.
const linkChangeGroups = unix.RTMGRP_LINK |
	unix.RTMGRP_IPV4_IFADDR |
	unix.RTMGRP_IPV6_IFADDR

// LinkChanges opens an rtnetlink multicast subscription and emits a
// struct{} on out every time the kernel reports a link or address state
// change. Message contents are intentionally discarded: callers
// re-fingerprint the host DNS environment on each settled event, so we
// only need the "something moved" signal. Returns on done close, or
// silently on socket open / bind failure so the caller falls back to
// time-based polling without a hard error.
func LinkChanges(out chan<- struct{}, done <-chan struct{}) {
	fd, err := unix.Socket(unix.AF_NETLINK,
		unix.SOCK_RAW|unix.SOCK_CLOEXEC|unix.SOCK_NONBLOCK,
		unix.NETLINK_ROUTE)
	if err != nil {
		return
	}
	defer unix.Close(fd)

	if err := unix.Bind(fd, &unix.SockaddrNetlink{
		Family: unix.AF_NETLINK,
		Groups: linkChangeGroups,
	}); err != nil {
		return
	}

	go func() {
		<-done
		_ = unix.Shutdown(fd, unix.SHUT_RDWR)
	}()

	buf := make([]byte, 4096)
	for {
		n, err := unix.Read(fd, buf)
		if err != nil {
			if errors.Is(err, unix.EAGAIN) || errors.Is(err, unix.EWOULDBLOCK) {
				if waitErr := waitReadable(fd, done); waitErr != nil {
					return
				}
				continue
			}
			return
		}
		if n <= 0 {
			return
		}
		select {
		case out <- struct{}{}:
		default:
		}
	}
}

// waitReadable blocks until fd has data or done closes. The netlink socket
// is non-blocking so a clean shutdown can interrupt the read by closing
// the socket, but we still need to park between bursts without spinning.
func waitReadable(fd int, done <-chan struct{}) error {
	pfd := []unix.PollFd{{Fd: int32(fd), Events: unix.POLLIN}}
	for {
		_, err := unix.Poll(pfd, 1000)
		if err != nil {
			if errors.Is(err, unix.EINTR) {
				select {
				case <-done:
					return errors.New("done")
				default:
					continue
				}
			}
			return err
		}
		select {
		case <-done:
			return errors.New("done")
		default:
		}
		if pfd[0].Revents&(unix.POLLIN|unix.POLLERR|unix.POLLHUP) != 0 {
			return nil
		}
	}
}
