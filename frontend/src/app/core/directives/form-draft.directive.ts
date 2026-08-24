import { Directive, Input, OnInit, OnDestroy, inject, Optional } from '@angular/core';
import { ControlContainer, FormGroupDirective } from '@angular/forms';
import { debounceTime, Subject, takeUntil } from 'rxjs';

@Directive({
  selector: '[appFormDraft]',
  standalone: true
})
export class FormDraftDirective implements OnInit, OnDestroy {
  @Input('appFormDraft') draftKey!: string;

  private destroy$ = new Subject<void>();
  private formGroupDirective = inject(FormGroupDirective, { optional: true });
  private controlContainer = inject(ControlContainer, { optional: true });

  ngOnInit() {
    if (!this.draftKey) {
      console.warn('FormDraftDirective requires a unique draftKey.');
      return;
    }

    const form = this.formGroupDirective?.form || (this.controlContainer as any)?.control;
    
    if (!form) return;

    // Restore draft if exists
    if (typeof window !== 'undefined') {
      const draft = sessionStorage.getItem(this.draftKey);
      if (draft) {
        try {
          form.patchValue(JSON.parse(draft), { emitEvent: false });
        } catch (e) {
          console.error('Failed to parse form draft', e);
        }
      }
    }

    // Auto-save on changes
    form.valueChanges.pipe(
      debounceTime(500),
      takeUntil(this.destroy$)
    ).subscribe((val: any) => {
      if (typeof window !== 'undefined') {
        sessionStorage.setItem(this.draftKey, JSON.stringify(val));
      }
    });
  }

  ngOnDestroy() {
    this.destroy$.next();
    this.destroy$.complete();
  }

  clearDraft() {
    if (typeof window !== 'undefined') {
      sessionStorage.removeItem(this.draftKey);
    }
  }
}
