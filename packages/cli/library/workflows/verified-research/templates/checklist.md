# Validation Checklist — Verified Research

Pre-flight gate before declaring "done." Every box must be ticked.

If you cannot tick a box, you are not done. Either complete the missing work or explicitly mark the gap as accepted (with reasoning) in the open questions section of your unified report.

---

## Research completeness

- [ ] Searched the team's knowledge base BEFORE dispatching researchers
- [ ] Searched the ticket tracker for related work BEFORE dispatching researchers
- [ ] Identified all relevant authoritative team documents
- [ ] Read the most recent versions of those documents

## Evidence quality

- [ ] Every code claim has `file.ext:line` citation
- [ ] Every doc claim has an exact URL
- [ ] Inferred or assumed claims are labeled
- [ ] Drift-scope verification ran on internal-code findings
- [ ] Drift items are documented in place with corrections visible

## Cross-reference completeness

- [ ] If a team-authored authoritative doc exists, I cross-referenced it
- [ ] If my synthesis diverges from that doc, I wrote a separate alignment file
- [ ] If the team doc is more authoritative, I re-baselined my primary document
- [ ] I did NOT silently overwrite earlier analysis

## Recommendation quality

- [ ] Answers the original questions directly (not tangentially)
- [ ] Trade-offs are explicit, not buried
- [ ] Alternatives considered are listed with rejection reasons
- [ ] Open questions are real, not handwavy "TBD"
- [ ] Adversarial review was run with 3+ rounds

## Existing-work integration

- [ ] Linked all related tickets and epics
- [ ] Linked all related wiki pages
- [ ] Identified work already in flight that overlaps with the recommendation
- [ ] Did not redo design work that's already been done elsewhere

## Contribution quality (if contributing back)

- [ ] Tone matches the source doc (verified by re-reading)
- [ ] Contributions are append-only
- [ ] Drop-in content is separate from rationale
- [ ] Review path is suggested
- [ ] Open questions for reviewer are explicit and friendly

## Output hygiene

- [ ] One findings file per research track
- [ ] One unified synthesis report
- [ ] One alignment file if external cross-reference happened
- [ ] One contribution file if contributing back
- [ ] All files in the agreed location
- [ ] Folder follows repo conventions

---

## Anti-pattern detector

If any of these are true, you have a quality problem to fix BEFORE declaring done:

- [ ] I skipped Phase 0 existing-work search "because the topic seemed niche"
- [ ] I skipped Phase 4 drift-scope "because the findings looked obviously correct"
- [ ] I skipped Phase 5 cross-reference "because I didn't find a team doc on first search"
- [ ] I overwrote my earlier synthesis instead of writing a new file when re-baselining
- [ ] I changed someone else's content while "contributing back"
- [ ] I asserted any claim without a citation
- [ ] I declared "done" without running this checklist

Each "yes" answer above is a violation. Fix before shipping.
