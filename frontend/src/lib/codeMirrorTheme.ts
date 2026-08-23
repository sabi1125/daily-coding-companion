import { EditorView } from "@codemirror/view"
import { HighlightStyle, syntaxHighlighting } from "@codemirror/language"
import { tags } from "@lezer/highlight"

export const codeEditorTheme = EditorView.theme({
  "&": {
    backgroundColor: "var(--code-surface)",
    color: "var(--code-plain)",
    fontSize: "0.875rem",
    height: "100%",
  },
  ".cm-content": {
    fontFamily: "var(--font-mono)",
    caretColor: "var(--code-plain)",
  },
  ".cm-gutters": {
    backgroundColor: "var(--code-gutter)",
    color: "var(--text-tertiary)",
    border: "none",
  },
  ".cm-activeLine": { backgroundColor: "transparent" },
  ".cm-activeLineGutter": { backgroundColor: "transparent" },
  "&.cm-focused": { outline: "none" },
})

export const codeHighlightStyle = syntaxHighlighting(
  HighlightStyle.define([
    { tag: tags.keyword, color: "var(--code-keyword)" },
    { tag: [tags.function(tags.variableName), tags.function(tags.propertyName)], color: "var(--code-function)" },
    { tag: [tags.string, tags.special(tags.string)], color: "var(--code-string)" },
    { tag: tags.number, color: "var(--code-number)" },
    { tag: tags.comment, color: "var(--code-comment)", fontStyle: "italic" },
    { tag: [tags.punctuation, tags.bracket], color: "var(--code-punctuation)" },
    { tag: [tags.variableName, tags.propertyName, tags.operator], color: "var(--code-plain)" },
  ])
)
