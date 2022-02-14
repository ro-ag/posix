package posix

import (
	"fmt"
)

func (m ModeT) valid(t ModeT) bool {
	if m&t != 0 {
		return true
	}
	return false
}

//S_ISBLK  block special
func (m ModeT) S_ISBLK() bool { return m&S_IFMT == S_IFBLK }

//S_ISCHR char special
func (m ModeT) S_ISCHR() bool { return m&S_IFMT == S_IFCHR }

//S_ISDIR directory
func (m ModeT) S_ISDIR() bool { return m&S_IFMT == S_IFDIR }

//S_ISFIFO fifo or socket
func (m ModeT) S_ISFIFO() bool { return m&S_IFMT == S_IFIFO }

//S_ISREG regular file
func (m ModeT) S_ISREG() bool { return m&S_IFMT == S_IFREG }

//S_ISLNK symbolic link
func (m ModeT) S_ISLNK() bool { return m&S_IFMT == S_IFLNK }

//S_ISSOCK socket
func (m ModeT) S_ISSOCK() bool { return m&S_IFMT == S_IFSOCK }

func (x DevT) major() DevT {
	return x >> 24 & 0xff
}
func (x DevT) minor() DevT {
	return x & 0xffffff
}

func FilePermStr(perm ModeT, flags int) string {

	/* If FP_SPECIAL was specified, we emulate the trickery of ls(1) in
	   returning set-user-ID, set-group-ID, and sticky bit information in
	   the user/group/other execute fields. This is made more complex by
	   the fact that the case of the character displayed for this bits
	   depends on whether the corresponding execute bit is on or off. */

	ru := '-'
	if perm.valid(S_IRUSR) {
		ru = 'r'
	}
	wu := '-'
	if perm.valid(S_IWUSR) {
		wu = 'w'
	}
	xu := ' '
	if perm.valid(S_IXUSR) {
		if perm.valid(S_ISUID) && flags&FP_SPECIAL != 0 {
			xu = 's'
		} else {
			xu = 'x'
		}
	} else {
		if perm.valid(S_ISUID) && flags&FP_SPECIAL != 0 {
			xu = 's'
		} else {
			xu = '-'
		}
	}
	rg := '-'
	if perm.valid(S_IRGRP) {
		rg = 'r'
	}
	wg := '-'
	if perm.valid(S_IWGRP) {
		wg = 'w'
	}
	xg := ' '
	if perm.valid(S_IXGRP) {
		if perm.valid(S_ISGID) && flags&FP_SPECIAL != 0 {
			xg = 's'
		} else {
			xg = 'x'
		}
	} else {
		if perm.valid(S_ISGID) && flags&FP_SPECIAL != 0 {
			xg = 's'
		} else {
			xg = '-'
		}
	}
	ro := '-'
	if perm.valid(S_IROTH) {
		ro = 'r'
	}
	wo := '-'
	if perm.valid(S_IWOTH) {
		wo = 'w'
	}
	xo := ' '
	if perm.valid(S_IXOTH) {
		if perm.valid(S_ISVTX) && flags&FP_SPECIAL != 0 {
			xo = 's'
		} else {
			xo = 'x'
		}
	} else {
		if perm.valid(S_ISVTX) && flags&FP_SPECIAL != 0 {
			xo = 's'
		} else {
			xo = '-'
		}
	}

	return fmt.Sprintf("[%04o] %c%c%c%c%c%c%c%c%c", perm, ru, wu, xu, rg, wg, xg, ro, wo, xo)
}

func (sb *Stat_t) DisplayStatInfo() {
	fmt.Printf("File type:                ")
	switch sb.Mode & S_IFMT {
	case S_IFREG:
		fmt.Println("regular file")
	case S_IFDIR:
		fmt.Println("directory")
	case S_IFCHR:
		fmt.Println("character device")
	case S_IFBLK:
		fmt.Println("block device")
	case S_IFLNK:
		fmt.Println("symbolic (soft) link")
	case S_IFIFO:
		fmt.Println("FIFO or pipe")
	case S_IFSOCK:
		fmt.Println("socket")
	default:
		fmt.Println("unknown file type?")
	}
	fmt.Printf("Device containing i-node: major=%d   minor=%d\n", sb.Dev.major(), sb.Dev.minor())
	fmt.Printf("I-node number:            %d\n", sb.Ino)
	fmt.Printf("Mode:                     %o (%s)\n", sb.Mode, FilePermStr(ModeT(sb.Mode), 0))

	if sb.Mode.valid(S_ISUID | S_ISGID | S_ISVTX) {
		uid, gid, sicky := "", "", ""
		if sb.Mode.valid(S_ISUID) {
			uid = "set-UID "
		}
		if sb.Mode.valid(S_ISGID) {
			gid = "set-GID "
		}
		if sb.Mode.valid(S_ISVTX) {
			sicky = "sticky "
		}
		fmt.Printf("    special bits set:     %s%s%s\n", uid, gid, sicky)
	}

	fmt.Printf("Number of (hard) links:   %d\n", sb.Nlink)
	fmt.Printf("Ownership:                UID=%d   GID=%d\n", sb.Uid, sb.Gid)

	if sb.Mode.S_ISCHR() || sb.Mode.S_ISBLK() {
		fmt.Printf("Device number (st_rdev):  major=%d; minor=%d\n", sb.Rdev.major(), sb.Rdev.minor())
	}

	fmt.Printf("File size:                %d bytes\n", sb.Size)
	fmt.Printf("Optimal I/O block size:   %d bytes\n", sb.Blksize)
	fmt.Printf("512B blocks allocated:    %d\n", sb.Blocks)
}
