// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

/*
package mutex provides a named machine level mutex shareable between processes.
[godoc-link-here]

Mutexes have names. Each each name, only one mutex for that name can be
acquired at the same time, within and across process boundaries. If a
process dies while the mutex is held, the mutex is automatically released.

The Linux/MacOS implementation uses flock, while the Windows implementation
uses a named mutex. On Linux, we also acquire an abstract domain socket for
compatibility with older implementations.
*/
package mutex
