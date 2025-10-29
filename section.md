# Unix Manual Sections

Unix-like systems organise their manual (man) pages into numbered sections.  
Each section groups related topics so the pager can disambiguate commands,
system interfaces, configuration files, and reference material that share the
same name. The table below summarises the sections you are most likely to
encounter on Linux, macOS, and other POSIX platforms.

```
 Section  Title / Focus                       Typical Content
 -------  ----------------------------------  --------------------------------------------
   1      User Commands                        Executables intended for end users (e.g. ls)
   1p     POSIX User Commands                  Standardised SUS/POSIX versions of section 1
   1m     Maintenance Commands (Solaris)       Legacy admin tools on System V style systems
   2      System Calls                         Kernel entry points (open, read, fork)
   3      Library Functions                    C library and other shared library APIs
   3p     POSIX Library Interfaces             Portable SUS/POSIX variants of section 3 calls
   4      Special Files / Drivers              /dev devices, driver interfaces, pseudo-files
   5      File Formats and Conventions         Config files, binary formats, protocol specs
   6      Games and Demonstrations             Historical games, educational programs
   7      Miscellaneous Information            Macro packages, standards, char maps (e.g. regex)
   8      System Administration Commands       Tools for superusers (mount, systemctl, zfs)
   9      Kernel Internals (Linux-specific)    Device driver APIs, kernel subsystems
```

## How Sections Are Used

- **Disambiguation:** If multiple pages share a topic name, specify the section:
  `man 5 crontab` shows the file-format reference, while `man 1 crontab` shows
  the user command.
- **Search Paths:** Man page files are stored under directories that include
  the section number, such as `/usr/share/man/man1/` or `/usr/share/man/man5/`.
  Many distributions compress these files, yielding suffixes like `.1.gz`.
- **POSIX Suffixes:** Sections ending with `p` contain interfaces required by
  the POSIX/SUS standards. They largely mirror their non-`p` counterparts but
  document only the behaviour guaranteed by the standard.
- **Platform Variations:** BSD derivatives, Solaris, and other flavours may
  add or rename sections (for example, Solaris uses section 5 for miscellany
  and 4 for file formats). Always consult `man man` on the target system to
  confirm the exact layout.

## Discovering Available Sections

- `man man` gives the authoritative overview for the local system, including
  any vendor-specific sections.
- `man -k topic` (apropos) searches topics across all sections and reports the
  section number alongside each match.
- `whatis topic` returns a one-line summary with the section header, making it
  easy to learn where a particular page lives.

Understanding the section hierarchy helps tooling like this project select the
right man page, present accurate examples, and build references without
confusing similarly named entries.
