# Classic image and text annotations 001–012: native capability gap

`features/ux/classic/timeline/annotation-highlights.feature` combines
seven iPad image-annotator clauses with five persistent text-highlight clauses.
The pinned Piclaw 3.2.4 shipped source map contains
`src/components/image-annotator.ts` and `src/components/post-highlights.ts`;
the latter documents the `PATCH /post/:id/annotations` path. Gi's native
`web/src/components/post.ts` shows audience/priority/updated metadata badges,
search-term highlights, and a stored-image lightbox, **not** an editable
annotation toolbar, saved selection offsets, image crop/export workflow or
persistent highlight API. Source searches found no native
`image-annotator`/`post-highlights` module or matching post-annotations endpoint.
These are capability gaps, not a licence to mark adjacent badges or lightbox
controls as annotation acceptance.

| Frozen IDs | Required interaction | Gi status |
|---|---|---|
| `001`–`005`, `007` | iPad-only inline image annotator with tools, two-finger pinch exclusion, crop/reset, flattened PNG upload/preview/cancel, SVG rasterization. | No native annotator. No tagged positive journey or physical iPad probe. |
| `006` | Non-iPad activation opens lightbox rather than annotator. | Installed Piclaw 3.2.4 mounted Classic probe uses a disposable image post and emulated non-iPad touch identity; click/tap open its lightbox, not the annotator, with an unsent draft retained. Gi's **partial, untagged** `lightbox.spec.mjs` test uses a real non-iPad browser identity, stored image, trusted click/tap and no annotator UI or writes. The paths differ (`/media/:id` in Piclaw, `/api/media/:id/raw` in Gi). Gi has no iPad-positive annotator path, so this does not verify an iPad-vs-non-iPad gate or whole `006`. Adjacent `013`–`016` lightbox tags do not fill it. |
| `008`–`012` | Text selection toolbar/colors; saved selection snapshot and `textOffset`; PATCH persistence; fine-pointer near-selection and coarse-pointer docked placement. | No native persistent text-highlight interaction/API or tagged journey. Search-term highlighting and metadata badges are unrelated. |

`make test-piclaw-non-ipad-lightbox` passed **6/6** Chromium/WebKit
phone/tablet/desktop cases for the existing lightbox branch, using disposable
timeline and PNG responses. Gi's focused untagged non-iPad journey passed
separately across the same six projects. Piclaw's emulated non-iPad identity,
fixture media and browser tap do not establish an actual device gate or
physical touch. No iPad-positive annotation interaction, upload, crop, text
highlight persistence or backend contract was exercised. Source-map module
presence alone proves none of those paths. Frozen Gherkin and Gi production
code were unchanged; the audit confirms fidelity of the existing non-iPad
lightbox slice rather than asking Gi to port the missing annotation surface.
