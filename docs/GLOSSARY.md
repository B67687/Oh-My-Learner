# GLOSSARY.md — Oh-My-Learner

Key terms used throughout the project, from learning science concepts to implementation details. Newcomers can use this as a quick reference when reading the code or documentation.

The glossary covers three layers: cognitive science foundations (retrieval practice, spaced repetition, desirable difficulties), study strategies (blocking, interleaving, self-explanation), and Oh-My-Learner implementation details (subject packs, backlog forgiveness, streak tracking, SM-2).

---

## Glossary

**Active Recall**

The cognitive process of retrieving information from memory during a learning task. Instead of re-reading notes or highlighting text (both forms of passive review), you force your brain to produce the answer from scratch. This act of retrieval strengthens the neural pathway, making the memory more durable and easier to access in the future.

Research by Karpicke and Roediger (2008) demonstrated that active recall produces nearly twice the retention of repeated study alone. Their landmark study showed that students who practiced retrieval retained about 80% of material after a week, while students who simply re-studied retained only about 40%.

This is the core mechanism behind every card in Oh-My-Learner. When you run learn review, each card presents a prompt, and your job is to retrieve the answer before seeing it. The tool does not show you the answer first. It expects you to produce it. That moment of effortful retrieval is what drives learning.

**Backlog Forgiveness**

A safeguard in Oh-My-Learner that prevents review pile-up after missed study days. Spaced repetition tools typically show all overdue cards when you return. If you miss a week, you might face hundreds of cards. Most users respond by skipping the session entirely, which only makes the problem worse.

Backlog forgiveness caps the number of due cards shown in a single session to a configurable daily limit. The default is 50 cards per day. Cards beyond that cap stay queued for future days. They do not disappear. They bleed into subsequent days gradually so you never face an overwhelming backlog.

The limit resets each day. If you studied 50 cards yesterday and 250 are still due, you see another 50 today. Over the course of a week, the backlog dissolves into your normal schedule. This keeps the tool forgiving enough for real-world use where daily consistency is not always possible.

**Blocking**

A study strategy where all cards from one topic are reviewed together before moving to the next topic. In a blocked session, you might see ten consecutive questions about operating system memory management before switching to data structures.

Oh-My-Learner uses blocking for declarative knowledge (facts, definitions, terminology). The decision is research-backed. Studies by Rohrer and Taylor (2007) and others have shown that interleaving harms recall for factual material. When you need to learn definitions and terminology, seeing similar items in sequence helps you build a coherent mental category.

A blocked review session in Oh-My-Learner groups all cards sharing the same subject together. Within each subject, cards may still vary by template type. The key distinction from interleaving is that the topic does not switch between every card.

**Declarative Knowledge**

Factual information: definitions, terminology, concepts, and facts. In Oh-My-Learner, declarative cards are blocked by subject during review. Examples include "What is a page fault?" or "Define virtual memory" or "List the three conditions for deadlock."

This type of knowledge benefits from focused, blocked practice rather than mixed review. When learning facts, seeing related facts in sequence helps your brain build associative networks. Mixing in unrelated topics between each fact interferes with this process.

Subject packs specify which templates produce declarative vs. procedural cards through template metadata. The scheduler reads this classification and chooses the appropriate review strategy. A template tagged as declarative produces cards that will be blocked by subject during review sessions.

The distinction between declarative and procedural knowledge maps to different brain systems. Declarative memories are hippocampus-dependent and benefit from consolidation during sleep. Repeated retrieval practice strengthens these memory traces over time.

**Desirable Difficulties**

Learning conditions that feel harder in the moment but produce better long-term retention. The concept comes from Bjork (1994) and is foundational to how Oh-My-Learner structures its review sessions.

Interleaving is the classic example of a desirable difficulty. Mixing topics feels more difficult and confusing than studying one topic at a time. Students often report that interleaving is frustrating because they struggle to answer questions. But that struggle is the point. The effort of switching between problem types forces deeper processing, and retention improves as a result.

Oh-My-Learner applies desirable difficulties by interleaving procedural cards across subjects. The review session feels harder than a blocked session, but research shows the extra difficulty produces measurably better long-term outcomes. The tool intentionally makes review harder in ways that help you learn.

**Ease Factor (EF)**

A parameter in the SM-2 algorithm that controls how fast review intervals grow for a given card. Each card has its own ease factor, and it changes over time based on your recall history.

Higher EF values mean the interval between reviews increases more quickly after successful recalls. The range runs from 1.3 (minimum, slowest interval growth) to 2.5 (default starting value). A card with EF 2.5 might double its interval after a successful review. A card with EF 1.3 might only increase by 30%.

Each time you recall a card successfully, the EF stays the same or increases slightly. When you forget a card, the EF decreases, slowing future interval growth so the card appears more often. This is how SM-2 adapts to card difficulty. Hard cards get shorter intervals and appear more frequently.

**FSRS (Free Spaced Repetition Scheduler)**

A modern alternative to SM-2 that uses machine learning to optimize review intervals. Unlike SM-2 which relies on a fixed mathematical formula with four static parameters, FSRS learns optimal interval patterns from real review data across thousands of users.

The result is a system that typically requires 20-30% fewer reviews than SM-2 to achieve the same retention rate. This means you spend less time reviewing while remembering just as much. The efficiency gain comes from more accurate interval predictions tailored to your memory patterns.

FSRS is planned as a future backend for Oh-My-Learner. The migration would replace the SM-2 scheduler in core/scheduler.go while keeping the same card storage, review interface, and user experience. The change would be invisible to users except for fewer daily reviews.

**Interleaving**

A study strategy where cards from different topics are mixed together during a single review session. Instead of seeing ten consecutive questions about one topic, you see a sorting algorithm question, then a debugging question about pointers, then a concurrency question.

Research by Rohrer and Taylor (2007) and Rohrer (2012) showed that interleaving improves long-term retention for procedural knowledge with an effect size of g=0.42. The strategy works because mixing topics forces your brain to identify the correct approach for each problem, rather than mindlessly applying the same strategy to a block of similar items.

Oh-My-Learner applies interleaving to procedural knowledge cards during review. The scheduler draws from multiple subjects and multiple template types, mixing them together so you cannot rely on context cues to guess the answer. You have to recognize the problem type itself.

**Knowledge Type**

The classification of learning material as either declarative (factual) or procedural (skill-based). This distinction is central to Oh-My-Learner because it determines the review strategy for each card.

Declarative cards are blocked by subject. Procedural cards are interleaved across subjects. The scheduler reads the knowledge type from each card's template metadata and routes it to the appropriate review strategy.

This two-type system is a simplification of broader cognitive taxonomies. It deliberately collapses a spectrum into two categories because the empirical literature consistently shows that the blocking vs. interleaving decision depends primarily on this distinction. Subject pack authors specify the knowledge type for each template, and the scheduler handles the rest.

**Procedural Knowledge**

How-to knowledge: coding patterns, debugging techniques, algorithm implementation, and problem-solving steps. In Oh-My-Learner, procedural cards are interleaved across subjects during review.

Examples include "Trace the output of this recursive function" or "Find the bug in this code snippet" or "Implement a binary search" or "Walk through what happens when a TLB miss occurs." These cards test your ability to apply a technique, not just recall a fact. They often present a novel scenario where you must transfer knowledge to an unfamiliar context.

This type of knowledge benefits from mixed practice because applying the right technique to the right problem is itself a skill that interleaving trains. When you practice identifying which technique fits each problem, you get better at that identification in real situations. A blocked session does not exercise this meta-cognitive skill.

Procedural knowledge in Oh-My-Learner uses template types like code-trace (follow the execution path), debug-find (identify the bug), and explain-why (reason about behavior). These formats test application and analysis, not just recall.

**Retrieval Practice**

The act of actively recalling information from memory during learning. It is the single highest-evidence learning technique available, according to the comprehensive meta-analysis by Dunlosky et al. (2013), which rated it as having "high utility" for learners of all ages and skill levels.

Retrieval practice outperforms re-reading, highlighting, summarization, and almost every other common study technique. The effect is robust across domains (science, history, vocabulary, problem-solving), age groups (elementary through professional), and testing formats (free recall, cued recall, multiple choice).

Every review session in Oh-My-Learner is a retrieval practice session. You see a prompt, you retrieve the answer from memory, and you rate your confidence. The act of retrieval is what strengthens the memory, not the act of seeing the answer. The tool is designed so that the majority of your time is spent retrieving, not reading.

**Self-Explanation**

After answering a card, the practice of explaining why the answer is correct. This technique has an effect size of g=0.55 according to Fiorella and Mayer (2015).

The process forces you to connect the specific fact or technique to your broader understanding. When you explain why a recursive function returns a particular value, you are not just confirming the answer. You are integrating that case into your mental model of recursion.

Oh-My-Learner prompts users for a self-explanation after each card during review. The prompt is simple: "Explain why your answer is correct." Even a brief one-sentence explanation significantly improves retention compared to answering alone. Users can type their explanation or skip it, but the prompt is there as a reminder to engage with the material at a deeper level.

**SM-2**

The spaced repetition algorithm developed by Piotr Wozniak for the SuperMemo 2 application in 1987. It is one of the earliest and most influential spaced repetition algorithms, still widely used in flashcard applications today.

SM-2 calculates review intervals using three parameters per card: the repetition count (how many times you have answered correctly in a row), the ease factor (how fast intervals grow), and the quality of the last recall (rated 0-5). After each review, the algorithm updates these parameters and computes the next interval. A perfect recall (rating 4-5) extends the interval. A failed recall (rating 0-2) resets the repetition count and decreases the ease factor.

SM-2 is the algorithm currently implemented in Oh-My-Learner. The implementation lives in core/scheduler.go. The core package exposes a function that takes a review quality rating and the current card state, then returns updated SM-2 parameters. The algorithm is deterministic and runs entirely locally with no external dependencies.

**Spaced Repetition**

The practice of scheduling review sessions at increasing intervals based on how well you remember each item. Cards you struggle with appear more often. Cards you know well appear less often.

The concept was first formalized by Hermann Ebbinghaus in 1885 through his forgetting curve research. He showed that memory decays exponentially over time, but that each successful recall slows the rate of decay. Spaced repetition exploits this by timing each review to occur just before the memory would be forgotten.

This is one of the most evidence-backed learning optimizations available. Oh-My-Learner implements spaced repetition through the SM-2 algorithm in core/scheduler.go. After each review, the algorithm adjusts the card's interval based on your recall quality rating. Over time, each card develops a personalized review schedule that optimizes for long-term retention with minimal time investment.

**Streak**

The number of consecutive days you have completed your reviews. Oh-My-Learner tracks streaks as a motivation metric displayed in the learn status output.

The tool offers one to two days of forgiveness. If you miss a single day, your streak is not broken. This prevents demotivation from isolated missed days (travel, illness, busy periods) while still encouraging consistent daily practice. The streak counter shows both your current streak and your longest streak.

Streak tracking is visible alongside due counts and per-subject progress. The goal is to provide positive reinforcement for consistency without punishing occasional gaps. Research on habit formation shows that missing a single day does not significantly harm habit development, but breaking a long streak can demotivate users enough that they quit entirely. The forgiveness window addresses this.

**Subject Pack**

A TOML file that defines a learning subject for Oh-My-Learner. It contains templates, template metadata, variable lists, prerequisite relationships, and knowledge type classifications.

Each template within a pack specifies a question format (standard Q&A, code trace, debug-find, explain-why), the knowledge type classification, and the variable substitutions that generate unique practice problems. Variables are defined with lists of possible values, and the template engine randomly selects from these lists each time a card is generated. This means the same subject pack produces different questions each session.

Subject packs can be hand-written by contributors or AI-generated through the agent package. They live in the subjects/ directory and are installed with the learn add command. A well-designed subject pack produces randomized practice problems each session through its parameterized template variables. The pack format is documented in the project's subject pack examples and specification files.

---

## References

The definitions in this glossary draw from the following sources:

- Bjork, R. A. (1994). Memory and metamemory considerations in the training of human beings. In Metacognition: Knowing about knowing.
- Dunlosky, J., Rawson, K. A., Marsh, E. J., Nathan, M. J., and Willingham, D. T. (2013). Improving students' learning with effective learning techniques: Promising directions from cognitive and educational psychology. Psychological Science in the Public Interest.
- Fiorella, L. and Mayer, R. E. (2015). Learning as a generative activity: Eight learning strategies that promote understanding. Cambridge University Press.
- Karpicke, J. D. and Roediger, H. L. (2008). The critical importance of retrieval for learning. Science.
- Rohrer, D. (2012). Interleaving helps students distinguish among similar concepts. Educational Psychology Review.
- Rohrer, D. and Taylor, K. (2007). The shuffling of mathematics problems improves learning. Instructional Science.
- Wozniak, P. (1987). Application of a computer to improve the results obtained in working with the SuperMemo method. (SM-2 algorithm.)
