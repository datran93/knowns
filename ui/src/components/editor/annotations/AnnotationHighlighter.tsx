import { useEffect, useRef, useState } from "react";
import { type Annotation } from "../../../hooks/useAnnotations";

interface AnnotationHighlighterProps {
	containerRef: React.RefObject<HTMLElement | null>;
	annotations: Annotation[];
	onAnnotationClick: (annotation: Annotation) => void;
}

export function AnnotationHighlighter({
	containerRef,
	annotations,
	onAnnotationClick,
}: AnnotationHighlighterProps) {
	const highlightLayerRef = useRef<HTMLDivElement>(null);
	const [ranges, setRanges] = useState<Array<{annotation: Annotation; range: Range}>>([]);

	// Build highlight markers whenever annotations or container content changes
	useEffect(() => {
		const container = containerRef.current;
		if (!container) return;

		const newRanges: Array<{annotation: Annotation; range: Range}> = [];

		// Use TreeWalker to find text nodes
		const walker = document.createTreeWalker(
			container,
			NodeFilter.SHOW_TEXT,
			{
				acceptNode: (node) => {
					// Skip code blocks, pre blocks, and annotation bubble elements
					const parent = node.parentElement;
					if (!parent) return NodeFilter.FILTER_REJECT;
					const tag = parent.tagName.toLowerCase();
					if (tag === "pre" || tag === "code" || parent.closest("[data-annotation-bubble]")) {
						return NodeFilter.FILTER_REJECT;
					}
					return NodeFilter.FILTER_ACCEPT;
				},
			}
		);

		for (const ann of annotations) {
			let walkerNode: Text | null = walker.currentNode as Text;
			while (walkerNode) {
				const text = walkerNode.textContent || "";
				const startIdx = text.indexOf(ann.quote);
				if (startIdx !== -1) {
					const range = document.createRange();
					range.setStart(walkerNode, startIdx);
					range.setEnd(walkerNode, startIdx + ann.quote.length);
					newRanges.push({ annotation: ann, range });
					break;
				}
				try {
					walkerNode = walker.nextNode() as Text | null;
				} catch {
					break;
				}
			}
			// Reset walker to start
			walker.currentNode = container;
		}

		setRanges(newRanges);
	}, [containerRef, annotations]);

	// Render highlight overlays
	useEffect(() => {
		const layer = highlightLayerRef.current;
		if (!layer) return;

		// Clear existing highlights
		layer.innerHTML = "";

		for (const { annotation, range } of ranges) {
			try {
				const rects = range.getClientRects();
			for (let i = 0; i < rects.length; i++) {
				const rect = rects[i];
				if (!rect || rect.width === 0) continue;
				const highlight = document.createElement("div");
					highlight.className = "annotation-highlight-marker";
					highlight.style.cssText = `
						position: absolute;
						background: ${annotation.color}55;
						border-bottom: 2px solid ${annotation.color};
						pointer-events: auto;
						cursor: pointer;
						box-sizing: border-box;
					`;
					highlight.dataset.annotationId = annotation.id;
					highlight.style.left = `${rect.left}px`;
					highlight.style.top = `${rect.top}px`;
					highlight.style.width = `${rect.width}px`;
					highlight.style.height = `${rect.height}px`;
					highlight.addEventListener("click", (e) => {
						e.stopPropagation();
						onAnnotationClick(annotation);
					});
					layer.appendChild(highlight);
				}
			} catch {
				// Range may be detached; skip
			}
		}
	}, [ranges, onAnnotationClick]);

	return (
		<div
			ref={highlightLayerRef}
			className="annotation-highlight-layer pointer-events-none fixed inset-0 z-[9998] overflow-hidden"
			style={{ pointerEvents: "none" }}
		/>
	);
}
