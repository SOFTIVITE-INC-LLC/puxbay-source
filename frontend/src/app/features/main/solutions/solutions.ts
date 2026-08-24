import { Component, ViewEncapsulation } from '@angular/core';

import { RouterModule } from '@angular/router';

@Component({
  selector: 'app-solutions',
  standalone: true,
  imports: [RouterModule],
  templateUrl: './solutions.html',
  encapsulation: ViewEncapsulation.None,
})
export class Solutions {}
