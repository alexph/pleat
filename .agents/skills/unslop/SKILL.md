---
name: unslop
description: "Fixes AI tells, must be used for all writing"
---

# Unslop

Apply to all authored prose, including replies, updates, documents, and comments. Draft or edit for the reader's purpose, then quietly review the rules below. Preserve meaning, voice, genre, and requested detail. Do not rewrite quotations, code, identifiers, or required templates for style. Return the requested output; explain edits only when asked or when a factual gap needs attention.

Each row specifies a trigger, repair, and exception. Defaults yield to deliberate, useful choices. Flags prompt judgment, never automatic replacement. Do not manufacture informality, mistakes, anecdotes, or sentence variation to seem human.

## Hard requirements

| Rule / trigger | Repair | Exception |
| --- | --- | --- |
| Invented evidence: unsupported facts, figures, anecdotes, or personal experience | Verify, remove, or identify the missing evidence; preserve supplied facts without adding specificity. | Fiction and hypothetical examples may contain invented details when clearly framed as such. Never present them as evidence. |
| Unfaithful quotation: edited or reconstructed words presented as exact | Check against the source; quote faithfully or label a paraphrase. | Marked omissions or editorial insertions may clarify without changing meaning. |
| Unsupported authority: invented citations, mismatched sources, “experts agree” | Verify that the source exists and supports this claim; name it, narrow the claim, or remove it. | Ordinary observations and explicitly labeled judgments need no decorative citation. |
| Overconfidence: guarantees, causation, or consensus beyond the evidence | Match certainty to evidence and state relevant conditions. | Definitive wording is appropriate for established facts or genuinely guaranteed behavior. |

## Default preferences

| Rule / trigger | Repair | Exception |
| --- | --- | --- |
| Empty opening: generic scene-setting before the subject | Lead with the subject, finding, or action. | Context may lead when needed to understand the point; narrative openings may serve the genre. |
| Inflated significance: “pivotal,” “transformative,” or sweeping legacy claims | State the supported consequence or delete the appraisal. | Explain substantial significance when evidence supports it. |
| Promotional tone: praise without observable substance | Describe capabilities, results, and relevant limitations. | Requested advertising may persuade, while keeping factual claims supportable. |
| Vague abstraction: unclear actors, actions, or changes; noun-heavy prose | Name who does what and what changes; turn nominalizations into verbs. | Keep precise abstractions and standard terminology needed by the audience. |
| Analysis without explanation: trailing “highlighting,” “reflecting,” or “ensuring” clauses | Give the mechanism or evidence, or delete the clause. | Keep the clause when it states a supported relationship. |
| Elaborate verbs: “serves as a demonstration of” | Use “shows,” “is,” or the most precise direct verb. | Retain wording that expresses a distinction the simpler verb loses. |
| Repetition: paragraphs or conclusions restating the same claim | Give each paragraph a distinct job; cut duplicate summaries. | Recaps may help long documents, teaching, or an explicitly requested summary. |
| Mechanical transitions: repeated “moreover” or “additionally” | State the relationship, such as cause, consequence, or qualification; omit needless connectors. | Keep transitions that orient the reader. |
| Canned ending: generic optimism or a compulsory wrap-up | End at the last useful fact, supported implication, decision, or next step. | A requested conclusion may synthesize established points. |
| Excessive hedging: stacked qualifiers | State uncertainty once and explain its reason when relevant. | Distinct uncertainties may each need qualification. |
| Excessive formatting: headings, bold labels, or bullets for every sentence | Use paragraphs; retain structure that aids navigation or comparison. | Procedures, reference material, and accessibility may need more structure. |
| Chat scaffolding: “Here's a polished version” inside finished prose | Deliver the artifact directly. | Brief framing may clarify alternatives, scope, or a consequential caveat. |

## Review flags

| Rule / trigger | Repair | Exception |
| --- | --- | --- |
| Stock vocabulary: clusters of “delve,” “tapestry,” “unlock,” “robust,” or similar clichés | Choose the precise ordinary word; delete empty modifiers. | Keep literal meanings, defined technical terms, and deliberate voice. |
| Stock contrast: “not X, but Y” with an invented foil | State Y directly. | Keep contrasts that correct a relevant misconception or distinguish real alternatives. |
| Forced symmetry: repeated triplets, identical sentence openings, or uniform paragraph shapes | Let the actual content determine structure; combine or cut redundant points. | Keep genuine three-part sets, parallel instructions, and intentional rhetorical patterns. |
| Excessive em dashes: repeated interruptions or dramatic pivots | Use simpler syntax, commas, parentheses, or separate sentences where clearer. | Keep a dash that usefully marks an interruption or emphasis. No punctuation quota. |

## Examples

Illustrative examples below are not factual claims about a real product.

- Before: “In today's rapidly evolving world, our groundbreaking platform unlocks seamless collaboration.” After: “The platform lets editors review the same draft.”
- Before: “The update marks a pivotal moment, highlighting its profound impact.” After: “The update lets users recover deleted drafts.” Use this repair only if that capability is established.
- Before: “This isn't about speed. It's about trust.” After: “Readers need to know where the figures came from.”
- Before: “It could potentially perhaps reduce delays.” After: “It may reduce delays; this has not been measured.”

Acceptable: “The result was statistically significant” when the analysis supports it; “The form saves a draft, not a submitted application” when correcting a likely misunderstanding; a three-step procedure with exactly three steps; an em dash marking a useful aside.

## Basis

These are editorial heuristics, not an authorship detector. Context matters more than isolated words.

- [Wikipedia: Signs of AI writing](https://en.wikipedia.org/wiki/Wikipedia:Signs_of_AI_writing): practical warning patterns, with caveats about attribution.
- [Reinhart et al., Do LLMs write like humans?](https://pmc.ncbi.nlm.nih.gov/articles/PMC11874169/): grammatical and rhetorical differences across genres; supports reviewing structure beyond vocabulary.
- [Geng and Trotta, Human-LLM Coevolution](https://aclanthology.org/2025.findings-acl.657/): changing word frequencies in academic abstracts; cautions against a static blacklist.
