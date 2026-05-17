import { useCallback, useEffect, useMemo, useState } from "react";

export interface Annotation {
	id: string;
	docPath: string;
	text: string;
	quote: string; // the highlighted text
	startOffset: number;
	endOffset: number;
	createdAt: number;
	updatedAt: number;
	color: string;
}

const STORAGE_KEY = "knowns-annotations";
const MAX_ANNOTATIONS = 500;

const DEFAULT_COLORS = [
	"#fef08a", // yellow
	"#bbf7d0", // green
	"#bfdbfe", // blue
	"#fecaca", // red
	"#e9d5ff", // purple
	"#fed7aa", // orange
	"#fbcfe8", // pink
];

function generateId(): string {
	return `ann_${Date.now()}_${Math.random().toString(36).slice(2, 9)}`;
}

function loadFromStorage(): Annotation[] {
	try {
		const raw = localStorage.getItem(STORAGE_KEY);
		if (!raw) return [];
		const parsed = JSON.parse(raw);
		return Array.isArray(parsed) ? parsed : [];
	} catch {
		return [];
	}
}

function saveToStorage(annotations: Annotation[]) {
	try {
		localStorage.setItem(STORAGE_KEY, JSON.stringify(annotations));
	} catch (e) {
		console.warn("Failed to save annotations to localStorage:", e);
	}
}

export function useAnnotations(docPath: string | null) {
	const [annotations, setAnnotations] = useState<Annotation[]>(() => loadFromStorage());

	// Sync with storage on mount and storage events (cross-tab)
	useEffect(() => {
		const handleStorage = (e: StorageEvent) => {
			if (e.key === STORAGE_KEY) {
				setAnnotations(loadFromStorage());
			}
		};
		window.addEventListener("storage", handleStorage);
		return () => window.removeEventListener("storage", handleStorage);
	}, []);

	// Filter to current doc
	const docAnnotations = useMemo(() => {
		if (!docPath) return [];
		return annotations.filter((a) => a.docPath === docPath);
	}, [annotations, docPath]);

	const addAnnotation = useCallback(
		(quote: string, text: string, startOffset: number, endOffset: number, color?: string) => {
			if (!docPath) return null;
			const now = Date.now();
			const newAnnotation: Annotation = {
				id: generateId(),
				docPath,
				text,
				quote,
				startOffset,
				endOffset,
				createdAt: now,
				updatedAt: now,
				color: color || DEFAULT_COLORS[annotations.length % DEFAULT_COLORS.length],
			};

			setAnnotations((prev) => {
				const updated = [newAnnotation, ...prev];
				if (updated.length > MAX_ANNOTATIONS) {
					// Remove oldest by createdAt
					updated.sort((a, b) => b.createdAt - a.createdAt);
					return updated.slice(0, MAX_ANNOTATIONS);
				}
				saveToStorage(updated);
				return updated;
			});

			return newAnnotation;
		},
		[docPath, annotations.length]
	);

	const updateAnnotation = useCallback((id: string, updates: Partial<Pick<Annotation, "text" | "color">>) => {
		setAnnotations((prev) => {
			const updated = prev.map((a) =>
				a.id === id ? { ...a, ...updates, updatedAt: Date.now() } : a
			);
			saveToStorage(updated);
			return updated;
		});
	}, []);

	const deleteAnnotation = useCallback((id: string) => {
		setAnnotations((prev) => {
			const updated = prev.filter((a) => a.id !== id);
			saveToStorage(updated);
			return updated;
		});
	}, []);

	const clearDocAnnotations = useCallback(
		(path: string) => {
			setAnnotations((prev) => {
				const updated = prev.filter((a) => a.docPath !== path);
				saveToStorage(updated);
				return updated;
			});
		},
		[]
	);

	return {
		annotations: docAnnotations,
		allAnnotations: annotations,
		addAnnotation,
		updateAnnotation,
		deleteAnnotation,
		clearDocAnnotations,
	};
}
