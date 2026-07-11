package cmd

import (
	"fmt"

	"github.com/B67687/Oh-My-Learner/core"
	"github.com/spf13/cobra"
)

var mapCmd = &cobra.Command{
	Use:   "map [subject]",
	Short: "Show dependency graph for subjects",
	Long: `Display the prerequisite relationship graph for installed subjects.

Without arguments, shows all subjects and their dependencies.
With a subject name, shows that subject's dependency tree.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := core.NewStorage(getDBPath())
		if err != nil {
			return fmt.Errorf("failed to open storage: %w", err)
		}
		defer store.Close()

		metas, err := store.SubjectMetas()
		if err != nil {
			return fmt.Errorf("failed to get subject map: %w", err)
		}

		if len(metas) == 0 {
			fmt.Println("No subjects installed.")
			return nil
		}

		// Build lookup
		nameByID := make(map[string]string)
		prereqsByID := make(map[string][]string)
		for _, m := range metas {
			nameByID[m.ID] = m.Name
			prereqsByID[m.ID] = m.Prerequisites
		}

		// Filter to one subject if specified
		filter := ""
		if len(args) > 0 {
			filter = args[0]
		}

		for _, m := range metas {
			if filter != "" && m.ID != filter && m.Name != filter {
				continue
			}

			fmt.Printf("\n  %s (%s)\n", m.Name, m.ID)
			if len(m.Prerequisites) == 0 {
				fmt.Println("    No prerequisites")
			} else {
				fmt.Println("    Requires:")
				for _, pid := range m.Prerequisites {
					pname := nameByID[pid]
					if pname == "" {
						pname = pid
					}
					fmt.Printf("      - %s (%s)\n", pname, pid)

					// Show transitive prerequisites (1 level)
					if transitive, ok := prereqsByID[pid]; ok && len(transitive) > 0 {
						for _, tpid := range transitive {
							tpname := nameByID[tpid]
							if tpname == "" {
								tpname = tpid
							}
							fmt.Printf("        - %s (%s)\n", tpname, tpid)
						}
					}
				}
			}

			// Show what depends on this subject (reverse deps)
			var dependents []string
			for _, other := range metas {
				if other.ID == m.ID {
					continue
				}
				for _, pid := range other.Prerequisites {
					if pid == m.ID {
						dependents = append(dependents, other.Name+" ("+other.ID+")")
						break
					}
				}
			}
			if len(dependents) > 0 {
				fmt.Println("    Required by:")
				for _, d := range dependents {
					fmt.Printf("      - %s\n", d)
				}
			}

			if filter != "" {
				// Show path from filter to its prerequisites recursively
				fmt.Println("    Dependency chain:")
				printChain(prereqsByID, nameByID, m.ID, "", m.Prerequisites, make(map[string]bool))
			}
		}

		return nil
	},
}

func printChain(prereqsByID map[string][]string, nameByID map[string]string, subjectID string, indent string, prereqs []string, visited map[string]bool) {
	for _, pid := range prereqs {
		if visited[pid] {
			fmt.Printf("%s    [!] %s (cycle detected)\n", indent, nameByID[pid])
			continue
		}
		pname := nameByID[pid]
		if pname == "" {
			pname = pid
		}
		fmt.Printf("%s    -> %s\n", indent, pname)
		visited[pid] = true
		if children, ok := prereqsByID[pid]; ok && len(children) > 0 {
			printChain(prereqsByID, nameByID, pid, indent+"  ", children, visited)
		}
	}
}
