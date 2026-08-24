import { Component, ViewEncapsulation } from '@angular/core';

import { RouterModule } from '@angular/router';

@Component({
  selector: 'app-careers',
  standalone: true,
  imports: [RouterModule],
  templateUrl: './careers.html',
  encapsulation: ViewEncapsulation.None,
})
export class Careers {}
