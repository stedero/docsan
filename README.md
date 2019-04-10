# docsan
Document Sanitizer for the new [IBFD](https://www.ibfd.org) [Tax Research Platform](https://research.ibfd.org)

The current [TRP](https://online.ibfd.org/kbase) (Tax Research Platform) presents many documents in HTML format. Most of these documents use JavaScript for rendering features such as generating an outline and looking up footer information from a back-end server.
The new TRP named TRP 3.0 is mainly written in Angular and all functionality that is needed to render documents will be part of the TRP 3.0 application. This makes logic in documents redundant. In the transitional period moving from the old TRP to TRP 3.0, documents should be properly rendered in both environments. Docsan makes that possible by transforming HTML documents on-the-fly to JSON, stripping all unwanted elements and JavaScript code and generating meta-tags in a JSON structure.

For testing and demos I deployed Docsan on [Heroku](https://docsan.herokuapp.com).

Kudos to [Flurin Egger](https://nl.linkedin.com/in/flurinegger) for the idea.
