import { createContext, useCallback, useContext, useMemo, useState } from "react";
import { useAnnotations, type Annotation } from "../hooks/useAnnotations";

interface AnnotationContextType {
	annotations: Annotation[];
	addAnnotation: (
		quote: string,
		text: string,
		startOffset: number,
		endOffset: number,
		color?: string
	) => Annotation | null;
	updateAnnotation: (id: string, updates: Partial<Pick<Annotation, "text" | "color">>) => void;
	deleteAnnotation: (id: string) => void;
	activeAnnotation: Annotation | null;
	setActiveAnnotation: (annotation: Annotation | null) => void;
	isEditing: boolean;
	setIsEditing: (editing: boolean) => void;
}

const AnnotationContext = createContext<AnnotationContextType | null>(null);

export function AnnotationProvider({
	children,
	docPath,
}: {
	children: React.ReactNode;
	docPath: string | null;
}) {
	const {
		annotations,
		addAnnotation,
		updateAnnotation,
		deleteAnnotation,
	} = useAnnotations(docPath);

	const [activeAnnotation, setActiveAnnotation] = useState<Annotation | null>(null);
	const [isEditing, setIsEditing] = useState(false);

	const value = useMemo(
		() => ({
			annotations,
			addAnnotation,
			updateAnnotation,
			deleteAnnotation,
			activeAnnotation,
			setActiveAnnotation,
			isEditing,
			setIsEditing,
		}),
		[annotations, addAnnotation, updateAnnotation, deleteAnnotation, activeAnnotation, isEditing]
	);

	return <AnnotationContext.Provider value={value}>{children}</AnnotationContext.Provider>;
}

export function useAnnotationContext() {
	const context = useContext(AnnotationContext);
	if (!context) {
		throw new Error("useAnnotationContext must be used within AnnotationProvider");
	}
	return context;
}

/** Safe version that returns null when outside AnnotationProvider instead of throwing. */
export function useAnnotationContextOptional() {
	return useContext(AnnotationContext);
}
