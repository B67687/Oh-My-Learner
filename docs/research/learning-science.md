# Research: Learning Science for University CS Study

**Generated:** 2026-07-11  
**Purpose:** Inform the evolution of `oh-my-learner` — a CLI template-based practice problem tool with SM-2 spaced repetition.  
**Scope:** What should the tool do, what should it change, and what should it leave alone.

---

## Summary (key findings for the tool)

The existing tool already nails the two **highest-utility** learning techniques (Dunlosky et al. 2013): **practice testing** (template-based problem generation) and **distributed practice** (SM-2 scheduler). That's the right foundation.

What's missing or underdeveloped for "super learner" support:

| Gap | Why It Matters | What to Add |
|-----|---------------|-------------|
| **Interleaving** (mixing subjects/problem types) | Second-most-supported technique after spaced repetition. Rohrer 2012, Bjork 1992. Already in VISION, not in scheduler. | `scheduler.go`: Implement interleaved session ordering. Mix due cards across subjects. |
| **Conceptual understanding** (not just facts) | Flashcards test recall, not understanding. Feynman technique, elaborative interrogation, self-explanation build deep knowledge. | Template types: "Explain why", "How does X relate to Y", "Find the bug", "Trace this code". |
| **Knowledge mapping** | CS concepts form dependency graphs. Students need to see how topics connect. | Subject packs declare prerequisites/related templates. `learn map` shows connections. |
| **Priority / what to study** | Overwhelming volume is the #1 student complaint. Pareto principle for learning. | Templates can have priority metadata. `learn suggest` surfaces high-leverage due items. |
| **Meta-learning / metacognition** | Effective learners plan, monitor, evaluate. Most students don't. | Post-session reflection prompts. `learn stats` with trend data. "Learning tips" in review flow. |
| **Streak mechanics** | Already in VISION as loss-aversion. Simple, effective. Expand slightly. | Visual streak counter. Weekly goal setting. "Don't break the chain" visualization. |
| **CS-specific exercises** | Programming = procedural knowledge. Debugging, tracing, code reading are separate skills. | Template categories: code-trace, debug-find, output-predict, design-compare. |

### What to leave alone

- **SM-2** → It's fine. FSRS is state-of-the-art but requires ~1000+ reviews to train and adds complexity. SM-2 works from day one. Consider FSRS only if user base grows and retention optimization matters more than simplicity.
- **No AI/LLM generation** → The VISION's anti-scope decision is correct. Deterministic templates are reliable, predictable, and debuggable.
- **No Anki import/export** → Correct. Different paradigms. Anki is card-first; oh-my-learner is template-first.
- **No cloud/accounts** → Correct. Local SQLite + CLI keeps it fast and private.

---

## Sources Surveyed

### Primary Research
1. Dunlosky, J., Rawson, K. A., Marsh, E. J., Nathan, M. J., & Willingham, D. T. (2013). *Improving Students' Learning With Effective Learning Techniques: Promising Directions From Cognitive and Educational Psychology.* Psychological Science in the Public Interest, 14(1), 4-58. — The definitive ranking of 10 study techniques.
2. Ye, J. et al. (2022). *A Stochastic Shortest Path Algorithm for Optimizing Spaced Repetition Scheduling.* Proceedings of the 28th ACM SIGKDD Conference. — FSRS algorithm publication.
3. Rohrer, D. (2012). *Interleaving helps students distinguish among similar concepts.* Educational Psychology Review, 24(3), 355-367.
4. Kornell, N. & Bjork, R.A. (2008). *Learning Concepts and Categories: Is Spacing the "Enemy of Induction"?* Psychological Science, 19(6), 585-592.
5. Roediger, H.L. & Karpicke, J.D. (2006). *Test-Enhanced Learning.* Psychological Science, 17(3), 249-255.
6. Bjork, R.A. (1994). *Memory and metamemory considerations in the training of human beings.* In J. Metcalfe & A. Shimamura (Eds.), Metacognition.
7. Brown, P.C., Roediger, H.L., & McDaniel, M.A. (2014). *Make It Stick: The Science of Successful Learning.* Harvard University Press. — Accessible synthesis of the research.

### Meta-Analysis & Reviews
8. Agarwal, P.K. et al. (2021). *A Meta-Analysis of Ten Learning Techniques.* Frontiers in Education, 6:581216. — Updated meta-analysis building on Dunlosky 2013.
9. Bisra, K. et al. (2018). *Inducing Self-Explanation: A Meta-Analysis.* Educational Psychology Review, 30(3), 703-725.

### Tools & Implementation
10. Anki documentation — FSRS integration (v23.10+), ease-hell problem analysis.
11. StudyGlen (2026). *Spaced Repetition Algorithms Explained: FSRS vs SM-2 vs Leitner.*
12. Mindomax (2026). *FSRS vs SM2 Spaced Repetition Algorithm.*
13. Gwern.net — *Spaced Repetition for Efficient Learning.* Comprehensive survey.

### Cognitive Science & Meta-Learning
14. Oakley, B. & Sejnowski, T. *Learning How to Learn* (Coursera). Focused vs. diffuse modes, chunking, procrastination.
15. MIT Teaching + Learning Lab — Metacognition resources.
16. Ertmer, P.A. & Newby, T.J. (1996). *The expert learner: Strategic, self-regulated, and reflective.* Instructional Science, 24(1), 1-24.

### Note-Taking & Knowledge Management
17. Forte, T. — *Progressive Summarization* (Forte Labs).
18. Cornell University — Cornell Note-Taking System (adapted 2026 guides).
19. Ahrens, S. (2017). *How to Take Smart Notes.* — Zettelkasten method.

---

## What Works (Ranked by Evidence)

Dunlosky et al. (2013) evaluated 10 study techniques on three criteria: (1) does it benefit learners of different ages/abilities, (2) does it work across different materials and tasks, (3) is the evidence base robust. The 2021 meta-analysis by Agarwal et al. (1,619 effects, 169,179 participants) confirmed the ranking with quantified effect sizes.

### HIGH UTILITY — Strong evidence, broad applicability

| Technique | Effect Size (Agarwal 2021) | Description |
|-----------|---------------------------|-------------|
| **Practice Testing (Active Recall)** | g = 0.74 | Self-testing, flashcards, practice problems. The act of retrieving information strengthens memory more than re-exposure. |
| **Distributed Practice (Spaced Repetition)** | g = 0.79 | Spreading study sessions over time rather than cramming. Dramatically improves long-term retention. |

**What the tool already does:** Both. Template generation = practice testing. SM-2 scheduler = distributed practice. This is the right foundation.

**How to strengthen:**
- Practice testing: Add more template *types* (not just Q&A). See CS-Specific section.
- Distributed practice: The SM-2 scheduler is sound. Consider a "cram mode" override (for pre-exam review).

### MODERATE UTILITY — Good evidence, narrower applicability

| Technique | Effect Size | Description |
|-----------|-------------|-------------|
| **Elaborative Interrogation** | g = 0.68 | Asking "Why is this true?" "How does this work?" Prompts deep processing. |
| **Self-Explanation** | g = 0.57 | Explaining steps in problem-solving, clarifying why an answer is correct. |
| **Interleaved Practice** | g = 0.44 | Mixing different types of problems/subjects in a single session. |

**What the tool is missing:** All three moderate-utility techniques are absent or underdeveloped.

**How to add:**
- **Elaborative interrogation**: Add a template field `explanation_prompt` — after answering, the tool asks "Why is that the answer?" and shows an explanation or asks the user to type one.
- **Self-explanation**: For multi-step problems, templates could include intermediate prompts: "What's the first step? Why?"
- **Interleaving**: Already in VISION as a main feature. Implement it in `scheduler.go`. When building a review session, mix due cards from different subjects and template types. Research shows this improves discrimination between similar concepts.

### LOW UTILITY — Weak evidence, don't prioritize

| Technique | Description — Why It's Weak |
|-----------|---------------------------|
| **Summarization** | Only helps with training; students are bad at it spontaneously. |
| **Highlighting / Underlining** | Illusion of competence. Does not improve retention. |
| **Keyword Mnemonic** | Works only for vocabulary, limited transfer. |
| **Imagery for Text** | Works for concrete text only; minimal benefit for abstract concepts. |
| **Rereading** | The most common student strategy. Creates familiarity, not knowledge. Illusion of mastery. (Roediger & Karpicke 2006) |

**What the tool should NOT do:** Don't build features that encourage passive consumption (e.g. "view notes mode", "read through mode").

---

## Spaced Repetition: State of the Art

### SM-2 (current implementation)
- **Pros:** Simple (~80 lines), proven (1987-present), needs no training data, works from first review.
- **Cons:** Universal curves — every card gets the same schedule. No per-learner adaptation. Can produce "ease hell" where stuck cards cycle frequently with no recovery.
- **Rating scale:** 0 (blackout) to 5 (perfect). Interval = previous × easiness_factor. Reset on failure.
- **Verdict for oh-my-learner:** Keep it. It's the right choice for a CLI tool with deterministic templates.

### FSRS (Free Spaced Repetition Scheduler)
- **Published:** 2022 (Ye et al., KDD 2022). Adopted by Anki as default in v23.10 (Nov 2023).
- **Model:** DSR — Difficulty (1-10), Stability (days to reach 90% recall), Retrievability (current recall probability).
- **Key advantage:** Per-learner parameter fitting via gradient descent. 20-30% fewer reviews for same retention vs SM-2 (based on benchmarks over 500M+ Anki reviews).
- **Requirement:** ~1,000+ reviews to fit meaningful personal parameters. Before that, falls back to defaults ≈ SM-2.
- **Verdict:** Overkill for oh-my-learner *today*. FSRS needs review history accumulation and a training mechanism. SM-2 is simpler, predictable, and matches the tool's philosophy of "no external deps." If adoption grows and users want retention optimization, FSRS can be added later — the scheduler interface (`ReviewCard`) makes this easy to swap.

### Leitner System
- Physical card system. 5 boxes with decreasing review frequency.
- Far coarser than SM-2/FSRS. Not relevant for a digital tool.
- **Verdict:** Ignore.

### Key Takeaway
The difference between SM-2 and FSRS is < the difference between using spaced repetition and not using it. SM-2 is adequate. Don't over-engineer.

---

## Knowledge Mapping Tools and Approaches

### The Problem
CS students struggle because concepts are deeply interconnected — understanding recursion requires understanding the call stack, which requires understanding function call semantics, which requires understanding memory. When you're lost, you don't know *which* prerequisite you're missing.

### Existing Tools

| Tool | Approach | Best For |
|------|----------|----------|
| **Obsidian** | Bi-directional links, graph view, Canvas (visual whiteboard) | Deep PKM, writing, research |
| **Logseq** | Outliner + bi-directional links at block level, PDF annotation | Task management + note-taking |
| **Roam Research** | Outliner with daily notes, block references | Journaling + networked thought |
| **InfraNodus** | AI-powered graph analysis of your notes | Gap detection in knowledge |
| **Mind maps** (Miro, Coggle, FreeMind) | Hierarchical radial diagrams | Brainstorming, planning |
| **Concept maps** (CmapTools) | Formal propositional structure with labeled arrows | Deep subject understanding |

### What's the Minimal Effective Approach for a CLI?

The tool shouldn't become a note-taking app. But it can express *relationships between problems*:

1. **Prerequisite metadata** — Each template can declare `prerequisites = ["complexity-sorting", "big-o-notation"]`. When a user struggles on a card, the tool can suggest: "This builds on Big O notation — review those templates first."
2. **Related concepts** — Templates can declare `related_to = ["heap-property", "sorting-property"]`. After answering, show related cards for interleaving.
3. **`learn map` command** — Render an ASCII dependency graph showing how topics in a subject connect. Simple edges: `A → B` (A is prerequisite for B).
4. **Weak signals** — Track which cards a user struggles with repeatedly. If card X is failed 3× and its prerequisite cards haven't been reviewed yet, surface that pattern.

### What NOT to do
- Don't build a graph visualization (leave that to Obsidian).
- Don't require manual linking — use TOML metadata in subject packs.
- Don't try to auto-infer relationships from content.

---

## Prioritization: What to Study First

### The Research

**Pareto Principle (80/20) for Learning:**
- ~20% of concepts in a subject account for ~80% of exam content and practical application.
- Foundational concepts (e.g., pointers in C, recursion, Big O) unlock vast amounts of subsequent material.
- A 2020 study from Johns Hopkins found the Pareto principle accelerates *initial* learning, but long-term expertise requires eventually covering the remaining 80%.

**Prerequisite Graphs:**
- Every CS course is structured as a dependency graph. You can't understand dynamic programming without recursion. You can't understand recursion without the call stack.
- "Minimum viable knowledge" — what's the smallest set of concepts you need to practice productively?

**Spacing Effect on Prioritization:**
- Cards that are most overdue (lowest retrievability) aren't always the highest leverage. A card seen 3 days ago at 90% retrievability is less important than a foundational concept seen 7 days ago at 60%.

### Recommendations for the Tool

1. **Priority metadata** — Templates have `priority = "high|medium|low"` or a numeric 1-5. High-priority items appear more frequently in interleaved sessions.
2. **Dependency-aware sessions** — `learn review --mode=foundations` shows only cards whose prerequisites are all mastered, ensuring the user never sees material they can't understand.
3. **Due-card triage** — `learn suggest` analyzes the due deck: "You have 42 due cards. The 12 priority-high ones cover graph algorithms, which is 33% of next week's exam. Start there."
4. **Mastery threshold** — A subject is "mastered" when all high-priority templates reach interval > 30 days. This gives a clear stopping signal.

---

## Existing Tool Landscape

### Anki (the incumbent)
- **Algorithm:** SM-2 → FSRS (default since v23.10)
- **Strengths:** Huge shared deck library, rich multimedia, mobile apps, mature ecosystem, extensive stats.
- **Weaknesses:** Card-writing burden (users must create every card manually), no template-based generation, poor support for interleaving (requires manual deck organization), no conceptual understanding exercises.
- **Why oh-my-learner is different:** Template generation removes card-writing friction. Interleaving is baked into the session model. CLI-first keeps it fast and focused.

### Mnemosyne
- **Algorithm:** SM-2 variant with simpler UI
- **Strengths:** Research-focused (used in studies), simple, open-source.
- **Weaknesses:** Fewer features, limited mobile support, no template generation.

### Org-drill (Emacs)
- **Algorithm:** SM-2
- **Strengths:** For users already in Emacs/Org-mode. Supports incremental reading.
- **Weaknesses:** Requires Emacs. Steep learning curve.

### SuperMemo
- **Algorithm:** SM-18/19 (proprietary, more advanced than FSRS per Woźniak's claims)
- **Strengths:** Decades of optimization, incremental reading, long-term tracking.
- **Weaknesses:** Proprietary, dated UI, Windows-only. Claims of outperforming FSRS are contested (SuperMemo tested unoptimized FSRS).

### What the Landscape Misses
1. **Template-based generation** — Only oh-my-learner does this as a core feature.
2. **Interleaving as first-class** — No tool makes mixing subjects/problem types an explicit feature.
3. **Conceptual understanding exercises** — No tool goes beyond flashcards to support Feynman technique, elaborative interrogation, or code-tracing exercises.
4. **CLI-first** — Every other tool is GUI-first. CLI offers speed, scriptability, and integration with coding workflows.

### Verdict
There is **no existing tool** that does what oh-my-learner aims to do. The landscape is all card-based SRS. Template-based interleaved practice is genuinely novel.

---

## Motivation and Habit Design

### The Research Base

**Streaks (Loss Aversion)**
- Kahneman & Tversky's prospect theory: losses hurt ~2× more than equivalent gains feel good.
- Duolingo's streak is their most effective engagement mechanic. Research (2025) shows streak rewards strengthen goal commitment.
- Already in VISION ✓. Keep it simple: `learn status` shows current streak. No points, badges, or leaderboards.

**Implementation Intentions**
- Gollwitzer (1999): "When situation X arises, I will perform behavior Y."
- Example: "When I finish my morning coffee, I will run `learn review` for 10 minutes."
- **Tool support:** `learn schedule` could prompt users to set a time/location trigger. Store in config as `review_trigger = "after morning coffee"`.

**Habit Stacking**
- Attach the new habit to an existing one. "After I brush my teeth, I review 5 cards."
- Related to implementation intentions. The tool could ask "What existing habit can you attach review to?" during setup.

**Variable Rewards**
- The most addictive apps (social media, gambling) use variable rewards — you don't know what you'll get.
- For a study tool: jumble the order of due cards, vary the mix of subjects, occasionally show "streak saver" cards (very easy ones you definitely know).
- **Specific suggestion:** `learn review` should randomize the order of due cards, not show them by interval. The unpredictability makes each session slightly novel.

**Intrinsic vs. Extrinsic Motivation**
- Deci & Ryan (Self-Determination Theory): Competence, autonomy, relatedness drive intrinsic motivation.
- Extrinsic rewards (streaks) can undermine intrinsic motivation if overdone. Use streaks as gentle accountability, not the main reason to study.
- **Tool implication:** Don't gamify beyond streaks. The VISION's anti-scope on points/badges/leaderboards is correct.

### Recommendations

1. **Keep streaks** (already in VISION). Show current streak length and longest streak.
2. **Add weekly goal setting:** `learn goal --reviews=50` — "Complete 50 reviews this week." Track progress.
3. **Post-session summary:** "You reviewed 12 cards (100% correct). Your streak is 5 days. You're on track for your weekly goal."
4. **Configurable review reminder:** A daily `learn` alias or shell integration could nudge. But don't send notifications (CLI tool — meet the user where they are).

---

## Meta-Learning: Becoming a Super Learner

### What Distinguishes Effective Learners

Research by Ertmer & Newby (1996), Oakley & Sejnowski, and the MIT Teaching + Learning Lab converges on these traits of expert learners:

| Trait | Description | Tool Support Possible? |
|-------|-------------|----------------------|
| **Metacognitive awareness** | Know what you know and don't know. Plan, monitor, evaluate. | Post-session reflection: "Which templates were hardest? What surprised you?" |
| **Strategy selection** | Match study technique to material type. | Tool could suggest technique: "This is a procedural topic — try tracing the code." |
| **Growth mindset** | Belief that ability is developed, not fixed. | Display progress over time. "You've mastered 30 templates — up from 15 last month." |
| **Chunking** | Build mental chunks from repeated practice. | SM-2 intervals naturally support this. |
| **Focused + Diffuse modes** | Alternating concentrated study with background processing. | Pomodoro timer integration? (Not core, but could wrap. Anti-scope? Questionable.) |
| **Deliberate practice** | Practice at the edge of competence with immediate feedback. | Templates provide immediate answer comparison. Difficulty should calibrate to just-right challenge. |

### The "Learning How to Learn" Course (Oakley/Sejnowski)

Key principles relevant to the tool:

1. **Focused vs. Diffuse Mode** — Brains need both concentrated attention and relaxed "background processing." The tool could suggest: "Study for 25 minutes, then take a 5-minute walk."
2. **Chunking** — Repeated practice builds mental chunks that free up working memory. The SM-2 schedule naturally creates this.
3. **Procrastination** — Oakley: use the Pomodoro technique (25 min focused work). The tool's CLI nature means low friction — `learn review` is one command, lowering the activation energy.
4. **Memory Palace / Visualization** — Not suitable for CLI. Skip.
5. **Metacognition** — "Test yourself before you feel ready." The tool inherently does this — the review session IS the test.

### Recommendations

1. **Meta-tips in review flow:** Occasionally insert a one-line learning tip between cards. "Did you know? Explaining a concept out loud (the Feynman technique) strengthens understanding more than re-reading."
2. **Progress trends in `learn stats`:** Show not just counts but trajectory — "Your retention rate has improved from 78% to 85% over the last 30 days."
3. **Post-session reflection prompt:** After review, ask "Rate how focused you felt (1-5)" and log it. Over time, show correlation between focus and performance.
4. **"Effective study methods" doc:** `learn guide` could display a cheatsheet of evidence-based techniques, sourced from this research.

---

## CS-Specific Learning Strategies

### Why CS Learning Is Different

**Procedural vs. Declarative Knowledge** — Most school subjects test declarative knowledge ("what is X"). Programming requires procedural knowledge ("how to do X"). Procedural knowledge is:
- Harder to acquire (requires practice, not reading)
- Harder to forget (once learned, like riding a bike)
- Best assessed by *doing*, not by recall

**Mental Models** — Programmers build internal simulations of how code executes. Novices have fragile mental models; experts have robust ones. A study by Sciencedirect (2023) synthesizes research on programmers' mental models: they guide all programming work and predict task performance.

**Debugging as Its Own Skill** — A 2024 ACM review found debugging requires five types of knowledge: domain (language), systems (program), procedural (how to debug), strategic (debugging strategies), and experiential (prior bug exposure). Most CS curricula teach none of these explicitly.

**The "Blocked Practice Trap"** — Students practice sorting algorithms → get good at sorting → can't distinguish when to sort vs. when to use a hash table. Interleaving *directly addresses* this by mixing problem types.

### Template Types for CS

The current `algorithms.toml` templates are entirely Q&A. CS learning needs more variety:

| Template Type | Example | Cognitive Skill |
|---------------|---------|-----------------|
| **Recall** (current) | "What is O(n log n) worst-case for comparison-based sorting?" | Declarative memory |
| **Code Trace** | "What does this code print? `for i in range(3): print(i)`" | Mental simulation, procedural |
| **Find the Bug** | "What's wrong with this function?" | Debugging, analytical |
| **Output Prediction** | "What is the output of `print(type([1,2,3]))`?" | Mental model of semantics |
| **Design Choice** | "Why use a hash table instead of a sorted array here?" | Design reasoning |
| **Compare & Contrast** | "What's the difference between a stack and a queue?" | Discriminative learning |
| **Apply Concept** | "Write pseudocode to check if a string is a palindrome." | Procedural generation |
| **Explain Why** | "Why does Bubble Sort stop early in the best case?" | Elaborative interrogation |
| **Incomplete Code** | "Fill in the missing line to complete this function." | Deliberate practice |

### Implementation for the CLI

The template format already supports variables. Adding a `type` field would let the tool adapt behavior:

```toml
[[templates]]
id = "trace-recursion"
type = "code-trace"         # NEW: type classification
question = "What does this print? def f(n): ... print(f(3))"
answer = "The output is 6."
explanation_prompt = "Walk through the call stack step by step."
```

- `type = "code-trace"` → show code with syntax highlighting (or ASCII emphasis).
- `type = "debug-find"` → present buggy code, ask user to identify the issue.
- `type = "design-compare"` → ask comparative question, show both sides.

### Debugging-Specific Features

The 2024 ACM review recommends deliberate practice with debugging:
- Templates with common bug patterns (off-by-one, null pointer, type error)
- Explanations that teach debugging strategies (print statements, binary search on code, rubber ducking)

---

## The Attention/Overwhelm Problem

### The Research

Students don't fail because they lack intelligence. They fail because they're overwhelmed by volume and don't know how to triage. Three note-taking/knowledge management systems address this:

**Progressive Summarization (Tiago Forte)**
- Layer 1: Original notes (capture everything)
- Layer 2: Bold passages (key points)
- Layer 3: Highlighted bold (most important)
- Layer 4: Executive summary (for the rest of us)
- Layer 5: Remix (original creation)
- **Relevance:** The tool already does layer-5 (practice problems). The concept of progressive layers could apply to template difficulty: start with core templates, unlock advanced ones.

**Cornell Notes**
- Cue column (questions), Notes column (facts), Summary (bottom).
- Forces interaction with notes twice: during capture and during cue review.
- **Relevance:** The Q&A format of templates is essentially Cornell cues. The tool could add a "summary mode" that groups recent cards by theme and shows aggregated answers.

**Zettelkasten (Niklas Luhmann)**
- Atomic notes (one idea per note). Bi-directional linking.
- Emergent understanding through connections, not hierarchy.
- **Relevance:** Each template is already atomic (one concept). Adding `related_to` and `prerequisites` metadata creates a lightweight Zettelkasten of practice problems.

### Addressing Overwhelm in the CLI

| Problem | Solution |
|---------|----------|
| "I have 200 due cards" | `learn review --limit=15` — always suggest a manageable session size. Default to 10-15 cards. |
| "I don't know what to study" | `learn suggest` — shows top priority items based on priority × overdue-ness. |
| "This subject feels endless" | `learn status` shows mastered vs. total. Clear progress signal. "You've mastered 12/31 templates." |
| "I keep failing the same card" | After 3 failures, suggest reviewing prerequisites. "You're struggling with heap operations — review tree properties first." |
| "Where do I start?" | `learn review --mode=foundations` — only high-priority items with mastered prerequisites. |

### Session Size Research

- Average optimal study session: 10-15 minutes (Oakley).
- Card count should be modest: 10-15 cards per session.
- The VISION's success criterion shows 5 cards in a session — that's actually too few for interleaving to work well. Aim for 10-15.

---

## Recommendations for oh-my-learner

### P0 — Implement This Week (Highest Impact / Lowest Effort)

1. **Interleaving in the scheduler** (`core/scheduler.go`)
   - When building a review session, mix due cards from different subjects and template types.
   - Research basis: Rohrer 2012, Bjork 1992, Dunlosky 2013 (moderate utility, but multiplies with spaced practice).
   - The VISION already promises this. The existing `DueCards` query returns cards ordered by time; the scheduler should shuffle them by subject/type.

2. **Template types** (`subjects/*.toml`)
   - Add an optional `type` field: `recall`, `code-trace`, `debug-find`, `design-compare`, `explain-why`.
   - The CLI can adapt display: code-trace shows code blocks, explain-why shows a "think about it" prompt before the answer.
   - Research basis: Elaborative interrogation (moderate utility), CS-specific skill development.

3. **`learn review --limit=N`**
   - Default session size 10-15 cards.
   - Prevents overwhelm and matches optimal study session length research.

### P1 — Implement This Month

4. **Prerequisite and related metadata**
   - Templates can declare: `prerequisites = ["template-id"]`, `related_to = ["template-id"]`.
   - `learn suggest` uses this to recommend foundational review.
   - Research basis: Knowledge mapping, concept dependency, Zettelkasten principles.

5. **Priority metadata**
   - Templates can declare `priority = 1-5` (5 = highest).
   - `learn suggest` and session building use priority to triage.
   - Research basis: Pareto principle for learning, minimum viable knowledge.

6. **`learn map` (ASCII dependency graph)**
   - Render templates and their prerequisite relationships in the terminal.
   - Simple text-based tree or graph. Not a visual masterpiece — just functional.

7. **Intra-session variation**
   - Randomize card order each session (variable rewards, novelty).
   - Vary difficulty mix within each session.

### P2 — This Semester

8. **Post-session reflection**
   - After each review session, show: "You got 11/15 correct. Your streak is 6 days. The template you struggled with most was 'heap-property'."
   - Optional: rate focus 1-5. Track trends.

9. **`learn stats` with trend data**
   - Show trajectory: retention rate over time, daily review count, streak history.
   - Gradual progress visualization supports growth mindset.

10. **"Struggle detection" — prerequisite suggestion on repeated failure**
    - If a card is failed 3+ times, check its prerequisites. If prerequisites haven't been reviewed recently, suggest: "This builds on template 'big-o-notation'. Review it first?"
    - Research basis: Knowledge mapping, Zettelkasten connection-finding.

11. **Debugging-specific template guidance**
    - For `type = "debug-find"`: show the buggy code, then after answer display a debugging strategy: "Try tracing the loop counter — what value does `i` have after the last iteration?"

### P3 — Long-term Consider

12. **FSRS as optional scheduler** — Only if users request it and review count exceeds 1000/session. Swap via interface.

13. **Weekly goal setting** — `learn goal --reviews=50` — Track completion rate. Gentle accountability.

14. **Implementation intentions setup** — During `learn init` or `learn config`, ask: "When will you review? What's your trigger?" Store it, show it.

### Never Do

- ❌ Points, badges, leaderboards, social features (extrinsic motivation undermines intrinsic)
- ❌ AI/LLM generation (unreliable, non-deterministic)
- ❌ Web UI, GUI, mobile app (scope creep)
- ❌ Anki import/export (different paradigm)
- ❌ Cloud sync or accounts (adds complexity, privacy risk)
- ❌ Note-taking features (leave that to Obsidian/Logseq)

---

## Quick Reference: Evidence Ratings for Decisions

| Decision | Rating | Source |
|----------|--------|--------|
| SM-2 is fine for a new tool | ✅ Keep | StudyGlen 2026, Mindomax 2026 |
| Interleaving should be default | ✅ Must add | Rohrer 2012, Bjork 1992, Dunlosky 2013 |
| Template types beyond Q&A | ✅ Must add | Elaborative interrogation research |
| Prerequisites in metadata | ✅ Add | Concept mapping literature |
| Priority system | ✅ Add | Pareto principle, curriculum sequencing |
| Streaks > gamification | ✅ Keep | Duolingo research, loss aversion |
| Pomodoro timer integration | 🤷 Optional | Oakley, weak evidence for toolification |
| FSRS over SM-2 | ❌ Not now | Needs 1000+ reviews to matter |
| AI/LLM generation | ❌ Never | VISION anti-scope, determinism |

---

## References

1. Dunlosky, J., Rawson, K. A., Marsh, E. J., Nathan, M. J., & Willingham, D. T. (2013). Improving Students' Learning With Effective Learning Techniques: Promising Directions From Cognitive and Educational Psychology. *Psychological Science in the Public Interest, 14*(1), 4-58. https://doi.org/10.1177/1529100612453266
2. Agarwal, P. K., et al. (2021). A Meta-Analysis of Ten Learning Techniques. *Frontiers in Education, 6*, 581216. https://doi.org/10.3389/feduc.2021.581216
3. Ye, J., et al. (2022). A Stochastic Shortest Path Algorithm for Optimizing Spaced Repetition Scheduling. *KDD '22*. https://doi.org/10.1145/3534678.3539081
4. Rohrer, D. (2012). Interleaving Helps Students Distinguish Among Similar Concepts. *Educational Psychology Review, 24*(3), 355-367.
5. Brown, P. C., Roediger, H. L., & McDaniel, M. A. (2014). *Make It Stick: The Science of Successful Learning*. Harvard University Press.
6. Roediger, H. L., & Karpicke, J. D. (2006). Test-Enhanced Learning. *Psychological Science, 17*(3), 249-255.
7. Bjork, R. A. (1994). Memory and metamemory considerations in the training of human beings. In J. Metcalfe & A. Shimamura (Eds.), *Metacognition*. MIT Press.
8. Oakley, B., & Sejnowski, T. *Learning How to Learn*. Coursera. https://www.coursera.org/learn/learning-how-to-learn
9. Ertmer, P. A., & Newby, T. J. (1996). The expert learner: Strategic, self-regulated, and reflective. *Instructional Science, 24*(1), 1-24.
10. Forte, T. Progressive Summarization. Forte Labs. https://fortelabs.com/blog/progressive-summarization-a-practical-technique-for-designing-discoverable-notes/
11. Ahrens, S. (2017). *How to Take Smart Notes*. CreateSpace.
12. StudyGlen (2026). Spaced Repetition Algorithms Explained. https://studyglen.com/guides/best-spaced-repetition-apps
13. Mindomax (2026). FSRS vs SM2 Spaced Repetition Algorithm. https://www.mindomax.com/fsrs-vs-sm2-spaced-repetition-algorithm
14. Structural Learning (2026). Interleaving: A Teacher's Guide. https://www.structural-learning.com/post/interleaving-a-teachers-guide
15. MIT Teaching + Learning Lab. Metacognition. https://tll.mit.edu/teaching-resources/how-people-learn/metacognition/
16. Kornell, N., & Bjork, R. A. (2008). Learning Concepts and Categories. *Psychological Science, 19*(6), 585-592.
