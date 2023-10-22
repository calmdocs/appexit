package appexit

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/imclaren/fileinfo/osext"

	//gops "github.com/keybase/go-ps"
	"github.com/shirou/gopsutil/v3/process"
)

// PID exits when process with pid exits
func PID(ctx context.Context, cancel context.CancelFunc, killPID *int) {
	if killPID != nil && *killPID != 0 {
		fmt.Println("Exit when the following process id (pid) exits:", *killPID)

		go func() {
			defer func() {
				cancel()
				os.Exit(0)
			}()

			idleDuration := 2 * time.Second
			idleDelay := time.NewTimer(idleDuration)
			defer idleDelay.Stop()
			for {
				idleDelay.Reset(idleDuration)
				select {
				case <-ctx.Done():
					return
				case <-idleDelay.C:
				}
				//p, err := gops.FindProcess(*killPID)
				//if err != nil || p == nil {
				ok, err := process.PidExistsWithContext(ctx, int32(*killPID))
				if err != nil || !ok {
					return
				}
			}
		}()
	}
}

// CloneWithLockFile exits when a new clone of this exe creates a new lock file at lockPath
// On startup, new clones create a lock file and then wait for 2 seconds for other clones to exit
// Once running, the exe checks for a lock file every second and exits if a lock file exists
func CloneWithLockFile(ctx context.Context, lockPath string) (err error) {

	// Delete lock file (if any) if no clones of this exe are running
	isClone, err := CheckIfClone(ctx)
	if err != nil {
		return err
	}
	if !isClone {
		os.Remove(lockPath)
	}

	// Exit if lock file exists
	// (i.e. an existing clone is running and we are in the 2 second lock wait time)
	_, err = os.Stat(lockPath)
	if err == nil {
		fmt.Println("start lock exit")
		os.Exit(0)
	}

	// Create a lock file and wait for clones to exit
	_, err = os.Create(lockPath)
	if err != nil {
		return err
	}
	defer os.Remove(lockPath)
	select {
	case <-ctx.Done():
		return
	case <-time.After(2 * time.Second):
	}
	os.Remove(lockPath)

	// Exit if a new clone starts
	go func() {
		idleDuration := 1 * time.Second
		idleDelay := time.NewTimer(idleDuration)
		defer idleDelay.Stop()
		for {
			_, err := os.Stat(lockPath)
			if err == nil {
				fmt.Println("lock exit")
				os.Exit(1)
			}
			idleDelay.Reset(idleDuration)
			select {
			case <-ctx.Done():
				return
			case <-idleDelay.C:
			}
		}
	}()
	return nil
}

// Clone exits if a clone of this exe is running
func Clone(ctx context.Context) (err error) {
	isClone, err := CheckIfClone(ctx)
	if err != nil {
		return err
	}
	if isClone {
		fmt.Println("exit - duplicate executable")
		os.Exit(1)
	}
	return nil
}

// CheckIfClone checks if a clone of this exe is running
func CheckIfClone(ctx context.Context) (isClone bool, err error) {
	exeName, err := osext.ExecutableName()
	if err != nil {
		return false, err
	}
	processList, err := process.ProcessesWithContext(ctx)
	if err != nil {
		return false, err
	}
	count := 0
	for _, p := range processList {
		p := p
		name, err := p.NameWithContext(ctx)
		if err != nil {
			if err.Error() == "invalid argument" {
				continue
			}
			return false, err
		}
		if name == exeName {
			count += 1
			if count >= 2 {
				return true, nil
			}
		}
	}
	return false, nil
}
